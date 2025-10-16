package hardware

import (
	"log"
	"math"
	"math/rand"
	"time"
)

type Slider struct {
	closeChan  chan string
	Value      float64
	UpdateChan chan float64
}

func NewSlider() *Slider {
	return &Slider{
		closeChan:  make(chan string),
		Value:      0.0,
		UpdateChan: make(chan float64),
	}
}

func (s *Slider) Init() error {
	log.Printf("Init Slider")
	return nil
}

func (s *Slider) Start() {
	go func() {
		defer close(s.closeChan)
		defer close(s.UpdateChan)
		for {
			select {
			case <-s.closeChan:
				break

			default:
				f, err := s.readValue()
				if err != nil {
					s.Value = -1.0
					// s.UpdateChan <- -1.0
					time.Sleep(time.Second * 5)
					continue
				}

				if math.Abs(s.Value-f) > 0.1 {
					s.Value = f
					log.Printf("value updated to %f", f)
					// s.UpdateChan <- f
				}
				time.Sleep(time.Second * 5)
			}
		}
		log.Println("slider stopped")
	}()
}

func (s *Slider) Stop() {
	log.Println("slider stopping")
	s.closeChan <- "close"
}

func (s *Slider) readValue() (float64, error) {
	return rand.Float64(), nil
}
