package main

import (
	"github.com/yasushisakai/ch-mechanical-table/hardware"
	"log"
)

func main() {
	portNames, err := hardware.PortLookup()

	if err != nil {
		log.Fatalf("%s", err)
	}

	log.Printf("slider: %s", portNames.Slider)
	log.Printf("button: %s", portNames.Button)
}
