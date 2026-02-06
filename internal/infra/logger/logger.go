package logger

import (
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/a2gx/weather/internal/infra/config"
)

// Logger обёртка над slog.Logger с возможностью закрытия ресурсов.
type Logger struct {
	*slog.Logger
	closer io.Closer
}

// New создаёт Logger на основе конфигурации.
// Устанавливает логгер как глобальный (slog.SetDefault).
func New(cfg config.LoggerConfig, service string) (*Logger, error) {
	level := parseLevel(cfg.Level)

	// Writer для вывода логов.
	// TODO: здесь можно расширить для async записи или отправки на удалённый сервер:
	//   - AsyncWriter с буферизацией (non-blocking запись)
	//   - io.MultiWriter для записи в несколько destinations
	//   - HTTP/gRPC writer для отправки на remote logging server
	var (
		w      io.Writer = os.Stdout
		closer io.Closer
	)

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: cfg.AddSource,
	}

	var h slog.Handler
	switch strings.ToLower(strings.TrimSpace(cfg.Format)) {
	case "text":
		h = slog.NewTextHandler(w, opts)
	case "", "json":
		h = slog.NewJSONHandler(w, opts)
	default:
		h = slog.NewJSONHandler(w, opts)
	}

	sl := slog.New(h).With(
		slog.String("service", service),
	)

	// Устанавливаем как глобальный логгер
	slog.SetDefault(sl)

	l := &Logger{
		Logger: sl,
		closer: closer,
	}

	return l, nil
}

// Close закрывает ресурсы логгера (если есть).
func (l *Logger) Close() error {
	if l.closer != nil {
		return l.closer.Close()
	}
	return nil
}

// parseLevel преобразует строку в slog.Level.
func parseLevel(s string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug
	case "info", "":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
