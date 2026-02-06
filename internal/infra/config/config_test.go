package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_DefaultConfig(t *testing.T) {
	workdir := t.TempDir()
	setWorkdir(t, workdir)

	writeFile(t, filepath.Join(workdir, "config", "default.yaml"), `http:
  host: 127.0.0.1
  port: 8080
`)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	assertHTTP(t, cfg, "127.0.0.1", 8080)
}

func TestLoad_ModeOverridesDefault(t *testing.T) {
	workdir := t.TempDir()
	setWorkdir(t, workdir)
	setEnv(t, EnvMode, "dev")

	writeFile(t, filepath.Join(workdir, "config", "default.yaml"), `http:
  host: 127.0.0.1
  port: 8080
`)
	writeFile(t, filepath.Join(workdir, "config", "dev.yaml"), `http:
  host: 0.0.0.0
  port: 9090
`)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	assertHTTP(t, cfg, "0.0.0.0", 9090)
}

func TestLoad_EnvOverridesFiles(t *testing.T) {
	workdir := t.TempDir()
	setWorkdir(t, workdir)
	setEnv(t, EnvMode, "dev")
	setEnv(t, EnvPrefix+"HTTP__HOST", "10.0.0.1")
	setEnv(t, EnvPrefix+"HTTP__PORT", "7070")

	writeFile(t, filepath.Join(workdir, "config", "default.yaml"), `http:
  host: 127.0.0.1
  port: 8080
`)
	writeFile(t, filepath.Join(workdir, "config", "dev.yaml"), `http:
  host: 0.0.0.0
  port: 9090
`)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	assertHTTP(t, cfg, "10.0.0.1", 7070)
}

func assertHTTP(t *testing.T, cfg *Config, host string, port int) {
	t.Helper()
	if cfg.Http.Host != host {
		t.Fatalf("Http.Host = %q, want %q", cfg.Http.Host, host)
	}
	if cfg.Http.Port != port {
		t.Fatalf("Http.Port = %d, want %d", cfg.Http.Port, port)
	}
}

func setWorkdir(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(dir, "config"), 0o755); err != nil {
		t.Fatalf("MkdirAll() error: %v", err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir() error: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(cwd)
	})
}

func setEnv(t *testing.T, key, value string) {
	t.Helper()
	if err := os.Setenv(key, value); err != nil {
		t.Fatalf("Setenv() error: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Unsetenv(key)
	})
}

func writeFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}
}
