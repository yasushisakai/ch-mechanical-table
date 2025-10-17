package main

import (
	"context"

	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	table "github.com/yasushisakai/ch-mechanical-table"
	"github.com/yasushisakai/ch-mechanical-table/hardware"

	"github.com/labstack/echo/v4"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
}

func main() {

	e := echo.New()

	portNames, err := hardware.PortLookup()

	if err != nil {
		log.Fatalf("failed to lookup hardware port: %s", err)
	}

	if portNames.Slider == "" {
		log.Fatalf("Slider missing")
	}

	if portNames.Button == "" {
		log.Fatalf("Button missing")
	}

	t := table.New(portNames)

	e.GET("/", t.WebSocketFunc)

	// CTRL-C
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	go func() {
		if err := e.Start(":8080"); err != nil && err != http.ErrServerClosed {
			log.Fatalf("shutting down echo server: %s", err)
		}
	}()

	if err := t.Start(); err != nil {
		log.Fatalf("error starting table: %s", err)
	}

	<-ctx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := t.Stop(ctx); err != nil {
		log.Fatal(err)
	}

	if err := e.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}

	log.Print("cooper hewitt table stopped")

}
