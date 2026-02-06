package main

import (
	"log/slog"
	"os"

	"github.com/a2gx/weather/internal/infra/config"
	"github.com/a2gx/weather/internal/infra/logger"
)

func main() {
	// 1. Загрузка конфигурации
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config load error", "error", err)
		os.Exit(1)
	}

	// 2. Инициализация логгера
	log, err := logger.New(cfg.Logger, "weather")
	if err != nil {
		slog.Error("logger init error", "error", err)
		os.Exit(1)
	}
	defer log.Close()

	log.Info("starting server",
		slog.String("env", cfg.Env),
		slog.String("host", cfg.Http.Host),
		slog.Int("port", cfg.Http.Port),
	)

	log.Debug("config loaded", slog.Any("config", cfg))

	// TODO: запуск HTTP сервера
}
