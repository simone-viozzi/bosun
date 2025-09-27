package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/simone-viozzi/bosun/internal/app"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 0) // no timeout by default
	defer cancel()

	a := app.New()
	if err := a.Run(ctx, os.Args); err != nil {
		log.Fatal(err)
	}
}
