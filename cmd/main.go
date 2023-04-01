package main

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/pressly/goose"

	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

var (
	Version = ""
)

func main() {
	logger := getLogger()
	logger.WithFields(logrus.Fields{
		"service_name":    "movies_service",
		"service_version": Version,
	}).Info("Starting service: ")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

	db := mustGetDb(ctx, logger, os.Getenv("DATABASE_URL"))
	mustApplyMigrations(logger, db, "misc/sql/migrations")

	handlers := &Handlers{db, logger}
	router := chi.NewRouter()
	router.Use(middleware.SetHeader("Content-Type", "application/json"))
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Heartbeat("/healthz"))
	router.Get("/movies", handlers.getMovies)

	s := http.Server{
		Addr:              fmt.Sprintf(":%s", os.Getenv("PORT")),
		Handler:           router,
		ReadTimeout:       time.Second * 5,
		ReadHeaderTimeout: time.Second * 5,
		WriteTimeout:      time.Second * 5,
		IdleTimeout:       time.Second * 5,
		BaseContext: func(listener net.Listener) context.Context {
			return ctx
		},
	}

	go func() {
		logger.Infof("Server is about start at http://localhost%s", s.Addr)
		if err := s.ListenAndServe(); err != nil {
			logger.Warnf("Server terminated: %v", err)
		}
	}()

	<-done
	tctx, tcancel := context.WithTimeout(context.Background(), time.Second*10)
	defer tcancel()
	logger.Info("Prepare for graceful termination")
	if err := s.Shutdown(tctx); err != nil {
		logger.Fatalf("an error occurred during shutdown: %v", err)
	}

	logger.Info("Server has been gracefully shutdown")
}

func getLogger() *logrus.Logger {
	logger := logrus.New()

	if os.Getenv("FLAG_DEBUG_MODE") == "enabled" {
		logger.SetLevel(logrus.DebugLevel)
	} else {
		logger.SetFormatter(&logrus.JSONFormatter{
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "time",
				logrus.FieldKeyLevel: "severity",
				logrus.FieldKeyMsg:   "message",
			},
		})
	}

	return logger
}

func mustGetDb(ctx context.Context, logger *logrus.Logger, url string) *sql.DB {
	db, err := sql.Open("postgres", url)
	if err != nil {
		logger.Error(err)
		logger.Fatalf("unable to open a connection to %s", url)
	}

	if err = db.PingContext(ctx); err != nil {
		logger.Error(err)
		logger.Fatalf("unable to ping %s", url)
	}

	return db
}

func mustApplyMigrations(logger *logrus.Logger, db *sql.DB, dir string) {
	goose.SetLogger(logger)
	if err := goose.SetDialect("postgres"); err != nil {
		logger.Fatalf("unable to set db dialect: %v", err)
	}

	workdir, err := os.Getwd()
	if err != nil {
		logger.Fatalf("unable to get workdir: %v", err)
	}

	if err = goose.Up(db, fmt.Sprintf("%s/%s", workdir, dir)); err != nil {
		logger.Fatalf("unable to apply database migrations: %v", err)
	}
}
