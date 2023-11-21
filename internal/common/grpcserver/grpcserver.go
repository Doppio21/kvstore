package grpcserver

import (
	"context"
	"fmt"
	"net"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type Config struct {
	Address string
}

type Dependencies struct {
	Log *logrus.Logger
}

type GRPCServer struct {
	cfg Config
	*grpc.Server
}

func InterceptorLogger(l *logrus.Logger) logging.Logger {
	return logging.LoggerFunc(func(_ context.Context, lvl logging.Level, msg string, fields ...any) {
		f := make(map[string]any, len(fields)/2)
		i := logging.Fields(fields).Iterator()
		if i.Next() {
			k, v := i.At()
			f[k] = v
		}
		l := l.WithFields(f)

		switch lvl {
		case logging.LevelDebug:
			l.Debug(msg)
		case logging.LevelInfo:
			l.Info(msg)
		case logging.LevelWarn:
			l.Warn(msg)
		case logging.LevelError:
			l.Error(msg)
		default:
			panic(fmt.Sprintf("unknown level %v", lvl))
		}
	})
}

func NewGRPCServer(cfg Config, deps Dependencies) *GRPCServer {
	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			logging.UnaryServerInterceptor(InterceptorLogger(deps.Log)),
		),
	)
	return &GRPCServer{
		Server: srv,
		cfg:    cfg,
	}
}

func (s *GRPCServer) Run(ctx context.Context) error {
	li, err := net.Listen("tcp", s.cfg.Address)
	if err != nil {
		return err
	}

	errCh := make(chan error)
	go func() {
		errCh <- s.Server.Serve(li)
	}()

	select {
	case <-ctx.Done():
		s.Server.GracefulStop()
	case err := <-errCh:
		return err
	}

	<-errCh
	return nil
}
