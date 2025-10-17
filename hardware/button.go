package hardware

import (
	"go.bug.st/serial"
	"log"
	"math/rand"
	"strings"
)

const BUTTON_VID string = "2341"
const BUTTON_PID string = "8037"

type Button struct {
	portName   string
	closeChan  chan struct{}
	DoneChan   chan struct{} // sends when I'm done stopping
	OutputChan chan float64
}

func NewButton(portName string) *Button {
	return &Button{
		portName:   portName,
		closeChan:  make(chan struct{}),
		DoneChan:   make(chan struct{}),
		OutputChan: make(chan float64),
	}
}

func (b *Button) Start() error {

	mode := &serial.Mode{
		BaudRate: 115200,
		DataBits: 8,
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
	}

	port, err := serial.Open(b.portName, mode)

	if err != nil {
		log.Printf("could not open button port: %s", err)
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
		}
	}()

	go func() {
		defer close(b.closeChan)
		defer close(b.OutputChan)
		for {
			select {
			case <-b.closeChan:
				b.DoneChan <- struct{}{}
				break

			case raw, ok := <-internalChan:
				if !ok {
					log.Println("Serial Internal Channel closed from the other side")
				}
				value := -1.0
				if strings.HasPrefix(raw, "SIN") {
					value = 1.0
				} else if strings.HasPrefix(raw, "DOU") {
					value = 2.0
				} else if strings.HasPrefix(raw, "TRI") {
					value = 3.0
				}
				b.OutputChan <- value
			}
		}
		log.Println("button stopped")
	}()

	return nil
}

func (s *Button) Stop() {
	s.closeChan <- struct{}{}
}

func (s *Button) readValue() (float64, error) {
	v := float64(rand.Intn(3))
	return v, nil
}
