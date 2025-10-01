package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/simone-viozzi/bosun/internal/app"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	a := app.New()
	if err := a.Run(ctx, os.Args); err != nil {
		log.Fatal(err)
	}
}
