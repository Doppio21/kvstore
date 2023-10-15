package main

import (
	"context"
	"kvstore/internal/manager"
	"kvstore/internal/server"
	"kvstore/internal/store/mapkv"
	"os/signal"
	"syscall"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

func main() {
	reg := prometheus.NewRegistry()

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		PadLevelText:  true,
	})

	log := logrus.StandardLogger()
	store := mapkv.NewStore()

	mgr := manager.New(manager.Config{
		UseCompression: true,
	}, manager.Dependencies{
		Store: store,
		Log:   log,
	})

	server := server.NewServer(server.Config{
		Address: "localhost:8080",
	}, server.Dependencies{
		Registry: reg,
		Manager:  mgr,
		Log:      log,
	})

	if err := server.Run(ctx); err != nil {
		log.Errorf("error: %v\n", err)
	}
}
