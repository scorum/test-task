package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jinzhu/configor"
	"github.com/scorum/account-svc/internal/config"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Server   config.ServerConfig `yaml:"server" required:"true"`
	DB       config.DBConfig     `yaml:"db" required:"true"`
	LogLevel string              `yaml:"log_level" default:"info"`
}

func main() {
	var (
		cfg    Config
		router = chi.NewRouter()
	)

	if err := configor.Load(&cfg, "configs/config.yml"); err != nil {
		logrus.WithError(err).Error("load config")
		return
	}

	server := setupServer(cfg.Server, router)

	go func() {
		logrus.Info("starting server")
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			logrus.WithError(err).Error("listen and serve")
		}
		logrus.Info("server stopped")
	}()

	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, syscall.SIGINT, syscall.SIGTERM)

	sig := <-terminate
	logrus.Infof("terminate: %q", sig)

	shutdownServer(&server)

	logrus.Info("done")
}

func setupServer(cfg config.ServerConfig, handler http.Handler) http.Server {
	return http.Server{
		Addr:         cfg.Addr,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout + 5*time.Second,
		Handler:      http.TimeoutHandler(handler, cfg.WriteTimeout, "request timeout"),
		ErrorLog:     log.New(logrus.StandardLogger().WriterLevel(logrus.ErrorLevel), "", 0),
	}
}

func shutdownServer(server *http.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logrus.WithError(err).Error("server shutdown")
	}
}
