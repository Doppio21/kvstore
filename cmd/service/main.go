package main

import (
	"context"
	"errors"
	"kvstore/internal/gateway"
	"kvstore/internal/storeservice"
	"os"
	"os/signal"
	"syscall"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "service",
		Usage: "Microservices for key-value storage",
		Commands: []*cli.Command{
			gateway.CLICommand,
			storeservice.CLICommand,
		},
	}

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := app.RunContext(ctx, os.Args); err != nil && !errors.Is(err, context.Canceled) {
		println(err.Error())
		os.Exit(1)
	}

	os.Exit(0)
}
