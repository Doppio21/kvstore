package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"kvstore/internal/manager"
	"kvstore/internal/store"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	shutdownTimeout = 5 * time.Second
)

type Config struct {
	Address string
}

type Dependencies struct {
	Manager manager.Manager
	Log     *logrus.Logger
}

type Server struct {
	cfg  Config
	deps Dependencies

	log *logrus.Entry
}

func NewServer(cfg Config, deps Dependencies) *Server {
	return &Server{
		cfg:  cfg,
		deps: deps,
		log:  deps.Log.WithField("component", "server"),
	}
}

func (s *Server) Run(ctx context.Context) error {
	var (
		mux = http.NewServeMux()
		srv = &http.Server{
			Addr:    s.cfg.Address,
			Handler: mux,
		}
	)

	mux.Handle("/", s.handleFunc())

	serverClosed := make(chan struct{})
	go func() {
		s.log.Info("server started")
		defer close(serverClosed)
		if err := srv.ListenAndServe(); err == nil && err != http.ErrServerClosed {
			s.log.Fatalf("listen and serve: %v", err)
		}
	}()

	select {
	case <-ctx.Done():
		s.log.Info("shutting down server gracefully")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown: %w", err)
		}
	case <-serverClosed:
	}

	s.log.Info("server finished")
	return nil

}

func (s *Server) handleFunc() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		key := strings.TrimPrefix(r.URL.Path, "/")
		s.log.Infof("request [Method: %s, Path: %s, Key :%s]",
			r.Method, r.URL.Path, key)

		switch r.Method {
		case http.MethodGet:
			res, err := s.deps.Manager.Get(r.Context(), store.Key(key))
			if err != nil && !errors.Is(err, manager.ErrNotFound) {
				w.WriteHeader(http.StatusInternalServerError)
				s.log.Errorf("failed to get key=%s: %v", key, err)
				return
			} else if errors.Is(err, manager.ErrNotFound) {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			data, err := json.MarshalIndent(&res, "", "  ")
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
			_, err = w.Write(data)
			if err != nil {
				s.log.Errorf("something wrong on write response: %v", err.Error())
			}

			return

		case http.MethodPut:
			if len(key) == 0 {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			value := make([]byte, 0, 1024)
			buff := make([]byte, 1024)
			for {
				n, err := r.Body.Read(buff)
				if err != nil && !errors.Is(err, io.EOF) {
					fmt.Println(err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				value = append(value, buff[:n]...)
				if errors.Is(err, io.EOF) {
					break
				}
			}

			err := s.deps.Manager.Set(r.Context(), store.Key(key), value)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
			return
		case http.MethodDelete:
			err := s.deps.Manager.Delete(r.Context(), store.Key(key))
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			return
		case http.MethodPost:
			res, err := s.deps.Manager.Scan(r.Context(), manager.ScanOptions{})
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			data, err := json.MarshalIndent(&res, "", "  ")
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
			_, err = w.Write(data)
			if err != nil {
				s.log.Errorf("something wrong on write response: %v", err.Error())
			}
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	})
}
