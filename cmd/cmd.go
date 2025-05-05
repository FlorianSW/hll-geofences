package main

import (
	"context"
	"github.com/floriansw/go-hll-rcon/rconv2"
	"github.com/floriansw/hll-geofences/data"
	"github.com/floriansw/hll-geofences/worker"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	level := slog.LevelInfo
	if _, ok := os.LookupEnv("DEBUG"); ok {
		level = slog.LevelDebug
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}))

	configPath := "./config.yml"
	if path, ok := os.LookupEnv("CONFIG_PATH"); ok {
		configPath = path
	}

	c, err := data.NewConfig(configPath, logger)
	if err != nil {
		logger.Error("config", err)
		return
	}

	defer func() {
		err = c.Save()
		if err != nil {
			logger.Error("save-config", err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	for _, server := range c.Servers {
		pool, err := rconv2.NewConnectionPool(rconv2.ConnectionPoolOptions{
			Logger:   logger,
			Hostname: server.Host,
			Port:     server.Port,
			Password: server.Password,
		})
		if err != nil {
			logger.Error("create-connection-pool", "server", server.Host, "error", err)
			continue
		}
		w := worker.NewWorker(logger, pool, server)
		go w.Run(ctx)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	logger.Info("graceful-shutdown")
	cancel()
}
