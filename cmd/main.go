package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/ryakadev/rdf-contrib-collector/config"
	"github.com/ryakadev/rdf-contrib-collector/internal/database"
	"github.com/ryakadev/rdf-contrib-collector/internal/logger"
	"github.com/ryakadev/rdf-contrib-collector/repository"
	"github.com/ryakadev/rdf-contrib-collector/transport"
	"github.com/ryakadev/rdf-contrib-collector/usecase"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	ctx := context.TODO()

	cfg := config.Load()
	level, err := zapcore.ParseLevel(cfg.LogLevel)
	if err != nil {
		log.Fatalf("Can't parse log level: %s", err)
	}

	zaplog := logger.New(level)
	db := database.NewConnection(ctx, cfg, zaplog)
	r := repository.New(db)
	u := usecase.New(r, zaplog)
	transport.NewTransport(cfg, zaplog, u)

	// Start http server
	fmt.Printf("Listening on %s:%d...\n", cfg.AppHost, cfg.AppPort)
	if err := http.ListenAndServe(fmt.Sprintf("%s:%d", "0.0.0.0", cfg.AppPort), nil); err != nil {
		log.Fatal("Can't start http server", zap.Error(err))
	}
}
