package app

import "context"

type App struct{}

func New() *App { return &App{} }

func (a *App) Run(ctx context.Context, args []string) error {
	// TODO: wire ports/adapters, parse config/flags, start services
	return nil
}
