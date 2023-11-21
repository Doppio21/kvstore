package storeservice

import (
	"context"
	"kvstore/internal/common/grpcserver"
	"kvstore/internal/storeservice/manager"
	"kvstore/internal/storeservice/server"
	"kvstore/internal/storeservice/store/badgerkv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Server  grpcserver.Config
	Manager manager.Config
	Store   badgerkv.Config
}

type Dependencies struct {
	Registry *prometheus.Registry
	Log      *logrus.Logger
}

type StoreService struct {
	cfg  Config
	deps Dependencies
}

func New(cfg Config, deps Dependencies) *StoreService {
	return &StoreService{
		cfg:  cfg,
		deps: deps,
	}
}

func (ss *StoreService) Run(ctx context.Context) error {
	store, err := badgerkv.New(ss.cfg.Store, badgerkv.Dependencies{
		Log: ss.deps.Log,
	})
	if err != nil {
		return err
	}

	mgr := manager.New(ss.cfg.Manager, manager.Dependencies{
		Store: store,
		Log:   ss.deps.Log,
	})

	srv := grpcserver.NewGRPCServer(ss.cfg.Server, grpcserver.Dependencies{
		Log: ss.deps.Log,
	})
	server.Register(server.Dependencies{
		Server:  srv.Server,
		Manager: mgr,
	})

	return srv.Run(ctx)
}
