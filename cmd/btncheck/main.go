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

	if portNames.Button == "" {
		log.Fatalf("button port is missing")
	}

	button := hardware.NewButton(portNames.Button)

	// CTRL-C
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	button.Start()
	defer button.Stop()

	for {
		select {
		case <-ctx.Done():
			break
		case buttonValue, ok := <-button.OutputChan:
			if !ok {
				log.Print("Button Update Channel was closed from other side")
				break
			}
			log.Printf("Button value: %f", buttonValue)
		}
	}
}
