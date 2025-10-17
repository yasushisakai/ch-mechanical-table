package hardware

import (
	"fmt"
	"go.bug.st/serial"
	"log"
	"math"
	"strconv"
	"strings"
	"time"
)

const SLIDER_VID string = "2341" // Arduino
const SLIDER_PID string = "0043"
const SLIDER_SERIAL_NUMBER string = "9563533313135161A211"

type Slider struct {
	portName   string
	closeChan  chan struct{}
	DoneChan   chan struct{} // sends when I'm done stopping
	Value      float64
	OutputChan chan float64
	InputChan  chan float64
}

func NewSlider(portName string) *Slider {
	return &Slider{
		portName:   portName,
		closeChan:  make(chan struct{}),
		DoneChan:   make(chan struct{}),
		Value:      0.0,
		OutputChan: make(chan float64),
		InputChan:  make(chan float64),
	}
}

func (s *Slider) Start() error {

	mode := &serial.Mode{
		BaudRate: 115200,
		DataBits: 8,
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
	}

	port, err := serial.Open(s.portName, mode)

	if err != nil {
		log.Printf("could not open slider port: %s", err)
		return err
	}

	internalChan := make(chan string)

	go func() {
		defer func() {
			if err = port.Close(); err != nil {
				log.Printf("could not close button port: %s", err)
			}
		}()

		buf := make([]byte, 128)
		cnt := 0
		for {
			n, err := port.Read(buf[cnt:])

			if err != nil {
				log.Printf("error reading from button: %s", err)
				break
			}
			cnt += n

			result := string(buf[:cnt])

			if strings.Count(result, "\n") < 1 {
				continue
			}

			internalChan <- result

			cnt = 0
			for i := range buf {
				buf[i] = 0
			}

			select {
			case f, ok := <-s.InputChan:
				if !ok {
					log.Println("Slider Input Channel have closed from other side")
					break
				}

				if f < 0.0 || 1.0 < f {
					log.Printf("Value out of range: ignoring")
					continue
				}

				mes := fmt.Sprintf("%0.3f\n", f)

				_, err := port.Write([]byte(mes))
				time.Sleep(time.Microsecond * 10)
				if err != nil {
					log.Fatalf("Sending value(%0.3f) failed: %s", f, err)
				}

			default: // no block for input
			}
		}
	}()

	go func() {
		defer close(s.closeChan)
		defer close(s.OutputChan)
		defer close(s.InputChan)

		for {
			select {
			case <-s.closeChan:
				s.DoneChan <- struct{}{}
				break

			case raw, ok := <-internalChan:
				if !ok {
					log.Println("Serial Internal Channel (Slider) close from the other side")
				}

				value, err := strconv.ParseFloat(strings.TrimSpace(raw), 8)
				if err != nil {
					// we ignore this
					log.Printf("raw: %s", raw)
					log.Println("Parse Failed: %s", err)
					continue
				}

				if math.Abs(value-s.Value) > 0.01 {
					s.Value = value
					s.OutputChan <- value
				}
			}
		}
		log.Println("slider stopped")
	}()

	return nil
}

func (s *Slider) Stop() {
	s.closeChan <- struct{}{}
}
