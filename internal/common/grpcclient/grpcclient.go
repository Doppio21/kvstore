package grpcclient

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Config struct {
	Address string
}

type Dependencies struct{}

type GRPCClient struct {
	cfg  Config
	deps Dependencies

	*grpc.ClientConn
}

func New(cfg Config, deps Dependencies) *GRPCClient {
	return &GRPCClient{
		cfg:  cfg,
		deps: deps,
	}
}

func (c *GRPCClient) Run(ctx context.Context) error {
	conn, err := grpc.DialContext(ctx, c.cfg.Address,
		grpc.WithBlock(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}

	c.ClientConn = conn
	return nil
}
