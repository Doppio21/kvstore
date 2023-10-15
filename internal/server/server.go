package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"kvstore/internal/manager"
	"kvstore/internal/store/kv"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

const (
	shutdownTimeout = 5 * time.Second
)

type Config struct {
	Address string
}

type Dependencies struct {
	Registry *prometheus.Registry
	Manager  manager.Manager
	Log      *logrus.Logger
}

type Server struct {
	cfg  Config
	deps Dependencies

	mc  *serverMetricsCollector
	log *logrus.Entry
}

func NewServer(cfg Config, deps Dependencies) *Server {
	return &Server{
		cfg:  cfg,
		deps: deps,
		mc:   newMetricColletor(),
		log:  deps.Log.WithField("component", "server"),
	}
}

func (s *Server) Run(ctx context.Context) error {
	s.deps.Registry.MustRegister(s.mc)

	router := gin.New()
	router.Use(LoggerMiddleware(s.log))
	router.Use(MetricsMiddleware(s.mc))

	router.GET("/:key", s.getHandler)
	router.PUT("/:key", s.setHandler)
	router.DELETE("/:key", s.deleteHandler)
	router.GET("/", s.scanHandler)
	router.GET("/metrics", gin.WrapH(promhttp.HandlerFor(s.deps.Registry, promhttp.HandlerOpts{})))

	var (
		srv = &http.Server{
			Addr:    s.cfg.Address,
			Handler: router,
		}
	)

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

func (s *Server) getHandler(c *gin.Context) {
	key := c.Param("key")
	res, err := s.deps.Manager.Get(c, kv.Key(key))
	if s.replyError(c, err) {
		return
	}

	resp := GetResponse{
		Key:   res.Key,
		Value: res.Value,
	}

	c.JSON(http.StatusOK, &resp)
}

func (s *Server) setHandler(c *gin.Context) {
	key := c.Param("key")
	if len(key) == 0 {
		c.Status(http.StatusBadRequest)
		return
	}

	value, err := io.ReadAll(c.Request.Body)
	if s.replyError(c, err) {
		return
	}

	err = s.deps.Manager.Set(c, kv.Key(key), value)
	if s.replyError(c, err) {
		return
	}
	c.Status(http.StatusOK)
}

func (s *Server) deleteHandler(c *gin.Context) {
	key := c.Param("key")
	err := s.deps.Manager.Delete(c, kv.Key(key))
	if s.replyError(c, err) {
		return
	}
	c.Status(http.StatusOK)
}

func (s *Server) scanHandler(c *gin.Context) {
	var (
		limit  int
		err    error
		prefix = c.Query("prefix")
	)

	if l := c.Query("limit"); l != "" {
		limit, err = strconv.Atoi(l)
		if s.replyError(c, err) {
			return
		}
	}

	res, err := s.deps.Manager.Scan(c, manager.ScanOptions{
		Limit:  limit,
		Prefix: prefix,
	})
	if s.replyError(c, err) {
		return
	}
	c.JSON(http.StatusOK, &res)
}

func (s *Server) replyError(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}

	resp := ErrorResponse{Message: err.Error()}
	code := http.StatusInternalServerError

	switch {
	case errors.Is(err, manager.ErrNotFound):
		code = http.StatusNotFound
	}

	c.JSON(code, &resp)
	return true
}
