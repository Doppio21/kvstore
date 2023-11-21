package server

import (
	"context"
	"kvstore/internal/protobuf/storepb"
	"kvstore/internal/storeservice/manager"

	"google.golang.org/grpc"
)

func Register(deps Dependencies) {
	storepb.RegisterStoreServer(deps.Server, &Server{
		deps: deps,
	})
}

type Dependencies struct {
	Server  *grpc.Server
	Manager manager.Manager
}

type Server struct {
	deps Dependencies

	storepb.UnimplementedStoreServer
}

func (s *Server) Get(ctx context.Context, req *storepb.GetRequest) (*storepb.GetResponse, error) {
	result, err := s.deps.Manager.Get(ctx, []byte(req.Key))
	if err != nil {
		return &storepb.GetResponse{
			Error: &storepb.Error{
				Message: err.Error(),
			},
		}, nil
	}

	return &storepb.GetResponse{
		Value: []byte(result.Value),
	}, nil
}

func (s *Server) Put(ctx context.Context, req *storepb.PutRequest) (*storepb.PutResponse, error) {
	err := s.deps.Manager.Set(ctx, []byte(req.Key), []byte(req.Value))
	if err != nil {
		return &storepb.PutResponse{
			Error: &storepb.Error{
				Message: err.Error(),
			},
		}, nil
	}

	return &storepb.PutResponse{
		Error: nil,
	}, nil
}
