package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"kvstore/internal/storeservice/client"
	"kvstore/internal/storeservice/manager"
	"net/http"
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
	Registry    *prometheus.Registry
	Log         *logrus.Logger
	StoreClient *client.Client
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
	value, err := s.deps.StoreClient.Get(c, key)
	if s.replyError(c, err) {
		return
	}

	resp := GetResponse{
		Key:   key,
		Value: string(value),
	}

	c.JSON(http.StatusOK, &resp)
}

func (s *Server) setHandler(c *gin.Context) {
	key := c.Param("key")
	if len(key) == 0 {
		c.Status(http.StatusBadRequest)
		return
	}

	val, err := io.ReadAll(c.Request.Body)
	if s.replyError(c, err) {
		return
	}

	err = s.deps.StoreClient.Put(c, key, val)
	if s.replyError(c, err) {
		return
	}

	c.Status(http.StatusOK)
}

func (s *Server) deleteHandler(c *gin.Context) {
	_ = c.Param("key")
	// TODO: GRPC request

	c.Status(http.StatusOK)
}

func (s *Server) scanHandler(c *gin.Context) {
	// var (
	// 	limit  int
	// 	err    error
	// 	prefix = c.Query("prefix")
	// )

	// if l := c.Query("limit"); l != "" {
	// 	limit, err = strconv.Atoi(l)
	// 	if s.replyError(c, err) {
	// 		return
	// 	}
	// }

	// TODO: GRPC request
	// c.JSON(http.StatusOK, &res)
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
