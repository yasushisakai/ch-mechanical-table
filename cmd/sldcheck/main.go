package main

import (
	"context"
	"github.com/yasushisakai/ch-mechanical-table/hardware"
	"log"
	"os"
	"os/signal"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
}

func main() {
	portNames, err := hardware.PortLookup()

	if err != nil {
		log.Fatalf("%s", err)
	}

	if portNames.Slider == "" {
		log.Fatalf("Slider port is missing")
	}

	slider := hardware.NewSlider(portNames.Slider)

	// CTRL-C
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	slider.Start()
	defer slider.Stop()

	for {
		select {
		case <-ctx.Done():
			break
		case sliderValue, ok := <-slider.OutputChan:
			if !ok {
				log.Print("Slider Update Channel was closed from other side")
				break
			}
			log.Printf("Slider value: %f", sliderValue)
		}
	}
}
