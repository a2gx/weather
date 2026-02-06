package logger

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	"github.com/a2gx/weather/internal/infra/config"
)

func TestNew(t *testing.T) {
	cfg := config.LoggerConfig{
		Level:  "info",
		Format: "json",
	}

	log, err := New(cfg, "test-service")
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	if log == nil {
		t.Fatal("New() returned nil logger")
	}
	if log.Logger == nil {
		t.Fatal("New() returned logger with nil slog.Logger")
	}
}

func TestNew_SetsDefaultLogger(t *testing.T) {
	cfg := config.LoggerConfig{
		Level:  "info",
		Format: "json",
	}

	log, err := New(cfg, "test-service")
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	// Проверяем, что глобальный логгер установлен
	if slog.Default() != log.Logger {
		t.Error("New() did not set default logger")
	}
}

func TestLogger_Close(t *testing.T) {
	cfg := config.LoggerConfig{
		Level:  "info",
		Format: "json",
	}

	log, err := New(cfg, "test-service")
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	// Close без closer должен возвращать nil
	if err := log.Close(); err != nil {
		t.Errorf("Close() returned error: %v", err)
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"DEBUG", slog.LevelDebug},
		{"  debug  ", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"INFO", slog.LevelInfo},
		{"", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"WARN", slog.LevelWarn},
		{"warning", slog.LevelWarn},
		{"WARNING", slog.LevelWarn},
		{"error", slog.LevelError},
		{"ERROR", slog.LevelError},
		{"unknown", slog.LevelInfo},
		{"invalid", slog.LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseLevel(tt.input)
			if got != tt.expected {
				t.Errorf("parseLevel(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestNew_JSONFormat(t *testing.T) {
	cfg := config.LoggerConfig{
		Level:  "info",
		Format: "json",
	}

	log, err := New(cfg, "test-service")
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	// Проверяем, что логгер создан (handler type проверить сложнее)
	if log == nil {
		t.Fatal("New() returned nil for json format")
	}
}

func TestNew_TextFormat(t *testing.T) {
	cfg := config.LoggerConfig{
		Level:  "info",
		Format: "text",
	}

	log, err := New(cfg, "test-service")
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	if log == nil {
		t.Fatal("New() returned nil for text format")
	}
}

func TestNew_DefaultFormat(t *testing.T) {
	cfg := config.LoggerConfig{
		Level:  "info",
		Format: "", // пустой формат -> json
	}

	log, err := New(cfg, "test-service")
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	if log == nil {
		t.Fatal("New() returned nil for empty format")
	}
}

func TestNew_UnknownFormat(t *testing.T) {
	cfg := config.LoggerConfig{
		Level:  "info",
		Format: "unknown", // неизвестный формат -> json
	}

	log, err := New(cfg, "test-service")
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	if log == nil {
		t.Fatal("New() returned nil for unknown format")
	}
}

func TestNew_ServiceAttribute(t *testing.T) {
	var buf bytes.Buffer

	cfg := config.LoggerConfig{
		Level:  "info",
		Format: "json",
	}

	// Создаём логгер с кастомным writer для проверки вывода
	log, err := newWithWriter(cfg, "my-service", &buf)
	if err != nil {
		t.Fatalf("newWithWriter() returned error: %v", err)
	}

	log.Info("test message")

	output := buf.String()
	if !strings.Contains(output, `"service":"my-service"`) {
		t.Errorf("expected service attribute in output, got: %s", output)
	}
}

func TestNew_LevelFiltering(t *testing.T) {
	var buf bytes.Buffer

	cfg := config.LoggerConfig{
		Level:  "warn",
		Format: "json",
	}

	log, err := newWithWriter(cfg, "test-service", &buf)
	if err != nil {
		t.Fatalf("newWithWriter() returned error: %v", err)
	}

	// Debug и Info должны быть отфильтрованы
	log.Debug("debug message")
	log.Info("info message")

	if buf.Len() != 0 {
		t.Errorf("expected no output for filtered levels, got: %s", buf.String())
	}

	// Warn должен пройти
	log.Warn("warn message")
	if !strings.Contains(buf.String(), "warn message") {
		t.Error("expected warn message in output")
	}

	buf.Reset()

	// Error тоже должен пройти
	log.Error("error message")
	if !strings.Contains(buf.String(), "error message") {
		t.Error("expected error message in output")
	}
}

func TestNew_JSONOutput(t *testing.T) {
	var buf bytes.Buffer

	cfg := config.LoggerConfig{
		Level:  "info",
		Format: "json",
	}

	log, err := newWithWriter(cfg, "test-service", &buf)
	if err != nil {
		t.Fatalf("newWithWriter() returned error: %v", err)
	}

	log.Info("test message", slog.String("key", "value"))

	// Проверяем, что вывод - валидный JSON
	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Errorf("output is not valid JSON: %v, got: %s", err, buf.String())
	}

	// Проверяем обязательные поля
	if result["msg"] != "test message" {
		t.Errorf("expected msg='test message', got: %v", result["msg"])
	}
	if result["service"] != "test-service" {
		t.Errorf("expected service='test-service', got: %v", result["service"])
	}
	if result["key"] != "value" {
		t.Errorf("expected key='value', got: %v", result["key"])
	}
	if result["level"] != "INFO" {
		t.Errorf("expected level='INFO', got: %v", result["level"])
	}
}

func TestNew_TextOutput(t *testing.T) {
	var buf bytes.Buffer

	cfg := config.LoggerConfig{
		Level:  "info",
		Format: "text",
	}

	log, err := newWithWriter(cfg, "test-service", &buf)
	if err != nil {
		t.Fatalf("newWithWriter() returned error: %v", err)
	}

	log.Info("test message", slog.String("key", "value"))

	output := buf.String()

	// Проверяем, что вывод содержит ожидаемые части
	checks := []string{"level=INFO", "test message", "service=test-service", "key=value"}
	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("expected output to contain %q, got: %s", check, output)
		}
	}
}

// newWithWriter — вспомогательная функция для тестов с кастомным writer.
func newWithWriter(cfg config.LoggerConfig, service string, w *bytes.Buffer) (*Logger, error) {
	level := parseLevel(cfg.Level)

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

	l := &Logger{
		Logger: sl,
		closer: nil,
	}

	return l, nil
}
