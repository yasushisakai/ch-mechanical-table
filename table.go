package table

import (
	"context"
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"golang.org/x/net/websocket"
	"io"
	"log"
	"strconv"
	"strings"
	"syscall"

	"github.com/yasushisakai/ch-mechanical-table/hardware"
	"github.com/yasushisakai/ch-mechanical-table/hub"
)

type Table struct {
	closeChan chan struct{}
	slider    *hardware.Slider
	button    *hardware.Button
	clientHub *hub.Hub
}

// Creates a new table. Needs to the port names of each hardware to start.
func New(pn hardware.PortNames) *Table {
	return &Table{
		closeChan: make(chan struct{}),
		slider:    hardware.NewSlider(pn.Slider),
		button:    hardware.NewButton(pn.Button),
		clientHub: hub.New(),
	}
}

// Starts the go routines for the hardware
// It also waits for the close chan signal to stop them
func (t *Table) Start() error {

	if err := t.slider.Start(); err != nil {
		mes := fmt.Sprintf("failed to start Slider: %s", err)
		return errors.New(mes)
	}

	if err := t.button.Start(); err != nil {
		mes := fmt.Sprintf("failed to start Button: %s", err)
		return errors.New(mes)
	}

	t.clientHub.Start()

	go func() {
		for {
			select {
			case <-t.closeChan:
				break
			case value, ok := <-t.slider.OutputChan:
				if !ok {
					log.Println("Slider Update Channel closed from other side")
					break
				}
				mes := hub.BroadcastMessage{
					Name:  "slider",
					Value: value,
				}
				t.clientHub.BroadcastChan <- mes

			case value, ok := <-t.button.OutputChan:
				if !ok {
					log.Println("Button Update Channel closed from other side")
					break
				}
				mes := hub.BroadcastMessage{
					Name:  "button",
					Value: value,
				}
				t.clientHub.BroadcastChan <- mes

			}
		}
	}()

	return nil
}

// Stops the go routines by sending a signal to closeChan.
func (t *Table) Stop(ctx context.Context) error {

	defer close(t.slider.DoneChan)
	defer close(t.button.DoneChan)
	defer close(t.clientHub.DoneChan)
	defer close(t.closeChan)

	t.closeChan <- struct{}{}
	t.slider.Stop()
	t.button.Stop()
	t.clientHub.Stop()

	sliderStopped := false
	buttonStopped := false
	hubStopped := false

	for {
		if sliderStopped && buttonStopped && hubStopped {
			log.Println("table stopped")
			return nil
		}
		select {
		case <-t.slider.DoneChan:
			sliderStopped = true
		case <-t.button.DoneChan:
			buttonStopped = true
		case <-t.clientHub.DoneChan:
			hubStopped = true
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// Handler function to be used in echo server
func (t *Table) WebSocketFunc(c echo.Context) error {

	s := websocket.Server{
        Handshake: func(config *websocket.Config, r *http.Request) error {
            // Completely disable origin checking (least secure):
            return nil
	},
	Handler: func(ws *websocket.Conn) {
		defer ws.Close()
		clientInChan := make(chan hub.BroadcastMessage)

		// Register
		t.clientHub.RegisterChan <- clientInChan
		defer func() { t.clientHub.UnregisterChan <- clientInChan }()

		// Input from client to me
		go func() {
			for {
				msg := ""
				err := websocket.Message.Receive(ws, &msg)
				if err != nil {
					if err == io.EOF {
						log.Printf("Client shut off")
						break
					}
					log.Printf("Error Recieving Message from client: %#v", err)
				}

				v, err := strconv.ParseFloat(strings.TrimSpace(msg), 8)
				if err != nil {
					log.Printf("Parsing to float value errored, ignoring message %q: %#v", msg, err)
					continue
				}
				t.slider.InputChan <- v
			}
		}()

		// Data from hardware to the client

		// send slider value once at the beginning
		mesString := fmt.Sprintf("%s: %f", "slider", t.Slider.Value)
		err := websocket.Message.Send(ws, mesString)
		if err != nil {
			if errors.Is(err, syscall.EPIPE) {
				log.Printf("broken pipe")
				return
			}
			log.Printf("error sending message: %#v", err)
			return
		}

		for {
			// Write
			mes, ok := <-clientInChan
			if !ok {
				log.Println("Client Channel closed from other side")
				return
			}

			mesString = fmt.Sprintf("%s: %f", mes.Name, mes.Value)
			err = websocket.Message.Send(ws, mesString)
			if err != nil {

				if errors.Is(err, syscall.EPIPE) {
					log.Printf("broken pipe")
					return
				}

				log.Printf("error sending message: %#v", err)
				return
			}
		}
	},
	}

	s.ServeHTTP(c.Response(), c.Request())
	return nil
}
