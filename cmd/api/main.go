package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jinzhu/configor"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/scorum/account-svc/internal/config"
	"github.com/scorum/account-svc/internal/db"
	"github.com/scorum/account-svc/internal/handler"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Server   config.ServerConfig `yaml:"server" required:"true"`
	DB       config.DBConfig     `yaml:"db" required:"true"`
	LogLevel string              `yaml:"log_level" default:"info"`
}

func main() {
	var cfg Config
	if err := configor.Load(&cfg, "configs/config.yml"); err != nil {
		logrus.WithError(err).Error("load config")
		return
	}

	conn, err := setupDBConnection(cfg.DB)
	if err != nil {
		logrus.WithError(err).Error("setup db connection")
		return
	}

	if err := migrateDB(conn); err != nil {
		logrus.WithError(err).Error("schema migrate")
		return
	}

	var (
		storage    = db.NewStorage(conn)
		apiHandler = handler.New(storage)
		server     = setupServer(cfg.Server, setupRouter(apiHandler))
	)

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

func setupRouter(h *handler.Handler) chi.Router {
	router := chi.NewRouter()
	router.Post("/v1/account/create", h.Create)
	router.Post("/v1/account/debit", h.Debit)
	router.Post("/v1/account/credit", h.Credit)
	router.Post("/v1/account/getBalance", h.GetBalance)
	return router
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

func setupDBConnection(cfg config.DBConfig) (*sqlx.DB, error) {
	if cfg.Addr == "" {
		return nil, nil
	}

	conn, err := sqlx.Open("postgres", cfg.Addr)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	conn.SetMaxOpenConns(cfg.MaxOpenConns)
	conn.SetMaxIdleConns(cfg.MaxIdleConns)

	if err := conn.Ping(); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}

	return conn, nil
}

func migrateDB(db *sqlx.DB) error {
	migrations := &migrate.FileMigrationSource{
		Dir: "migrations",
	}

	n, err := migrate.Exec(db.DB, "postgres", migrations, migrate.Up)
	if err != nil {
		return fmt.Errorf("sql-migrate: %w", err)
	}

	logrus.Infof("Applied %d migrations!\n", n)
	return nil
}
