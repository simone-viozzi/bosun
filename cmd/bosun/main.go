package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/simone-viozzi/bosun/internal/cmd"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	rootCmd := cmd.NewRootCmd()
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		log.Fatal(err)
	}
}
