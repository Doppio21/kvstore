package gateway

import (
	"context"
	"kvstore/internal/common/grpcclient"
	"kvstore/internal/gateway/server"
	"kvstore/internal/storeservice/client"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Server server.Config
	Client grpcclient.Config
}

type Dependencies struct {
	Registry *prometheus.Registry
	Log      *logrus.Logger
}

type Gateway struct {
	cfg  Config
	deps Dependencies
}

func New(cfg Config, deps Dependencies) *Gateway {
	return &Gateway{
		cfg:  cfg,
		deps: deps,
	}
}

func (gw *Gateway) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	cl := grpcclient.New(gw.cfg.Client, grpcclient.Dependencies{})
	if err := cl.Run(ctx); err != nil {
		return err
	}

	client := client.New(cl)
	server := server.NewServer(gw.cfg.Server, server.Dependencies{
		Registry:    gw.deps.Registry,
		Log:         gw.deps.Log,
		StoreClient: client,
	})

	return server.Run(ctx)
}
