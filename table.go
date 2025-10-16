package table

import (
  "log"
  "errors"
  "github.com/labstack/echo/v4"
  "golang.org/x/net/websocket"

  "github.com/yasushisakai/ch-mechanical-table/hardware"
)

type Table struct {
	closeChan chan string
	slider *hardware.Slider
}


// Creates a new table.
func New() *Table {
	return &Table{
		closeChan: make(chan string),
		slider: hardware.NewSlider(),
	}
}

// Initializes a table. Will try to set the right ports.
// Returns an error if the hardware parts could not find them.
func (t *Table) Init() error {

  log.Printf("Table Init()")
  if err := t.slider.Init(); err != nil {
   return errors.New("Failed to initialize table")
  }

  return nil
}

// Starts the go routines for the hardware
// It also waits for the close chan signal to stop them
func (t *Table) Start() {
	log.Printf("Table Start()")

	go func() {
		defer close(t.closeChan)
		t.slider.Start()

		<- t.closeChan
		t.slider.Stop()
	}()
}


// Stops the go routines by sending a signal to closeChan.
func (t *Table) Stop() {
	t.closeChan <- "close"
}

// Handler function to be used in echo server
func (t *Table) WebSocketFunc(c echo.Context) error {
	websocket.Handler(func(ws *websocket.Conn) {
		defer ws.Close()
		for {
			// Write
			err := websocket.Message.Send(ws, "Hello, Client!")
			if err != nil {
				c.Logger().Error(err)
			}

			// Read
			msg := ""
			err = websocket.Message.Receive(ws, &msg)
			if err != nil {
				c.Logger().Error(err)
			}
			log.Printf("%s\n", msg)
		}
	}).ServeHTTP(c.Response(), c.Request())
	return nil
}