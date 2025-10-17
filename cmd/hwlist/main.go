package main

import (
	"github.com/yasushisakai/ch-mechanical-table/hardware"
	"log"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
}

func main() {
	if err := hardware.ListPortDetails(); err != nil {
		log.Fatalf("error listing ports details")
	}
}
