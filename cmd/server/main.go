package main

import (
	"context"

	"net/http"
	"os"
	"os/signal"
	"time"

	table "github.com/yasushisakai/ch-mechanical-table"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

func main() {

	e := echo.New()
	e.Logger.SetLevel(log.INFO)

	t := table.New()
	if err := t.Init(); err != nil {
		e.Logger.Fatalf("Table initizalition failed: %s", err)
	}

	e.GET("/", t.WebSocketFunc)

	// CTRL-C
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	go func() {
		if err := e.Start(":8080"); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatalf("shutting down echo server: %s", err)
		}
	}()

	t.Start()

	<-ctx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	t.Stop()

	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}

	e.Logger.Print("cooper hewitt table stopped")

}
