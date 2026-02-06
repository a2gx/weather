package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type Config struct {
	Env string `koanf:"env"`

	Http struct {
		Host string `koanf:"host"`
		Port int    `koanf:"port"`
	} `koanf:"http"`
}

const (
	EnvMode   = "ENV_MODE"
	EnvPrefix = "APP_" // префикс env overrides, напр. APP_HTTP__ADDR=":8080"

	defaultConfigDir  = "config/"
	defaultConfigFile = "default"
)

func Load() (*Config, error) {
	// koanf instance (разделитель ключей ".")
	k := koanf.New(".")

	// 1.: base: config/default.yaml (обязательный)
	baseFile := fmt.Sprintf("%s/%s.yaml", defaultConfigDir, defaultConfigFile)
	if err := loadYamlFile(k, baseFile, true); err != nil {
		return &Config{}, err
	}

	// 2.: mode: config/<ENV_MODE>.yaml (не обязательный)
	mode := strings.TrimSpace(os.Getenv(EnvMode))
	if mode != "" {
		modeFile := filepath.Join(defaultConfigDir, fmt.Sprintf("%s.yaml", mode))
		_ = loadYamlFile(k, modeFile, false)
	}

	// 3.: env overlay (самый высокий приоритет)
	// Формат:
	// APP_HTTP__ADDR=":8080"
	// APP_POSTGRES__PASSWORD="secret"
	// "__" превращаем в "."
	if err := k.Load(env.Provider(".", env.Opt{
		Prefix: EnvPrefix,
		TransformFunc: func(k, v string) (string, any) {
			// Приводим ключи к lower и разворачиваем вложенность через "__".
			k = strings.ReplaceAll(strings.ToLower(strings.TrimPrefix(k, EnvPrefix)), "__", ".")
			// Для списков разрешаем пробелы как разделитель.
			if strings.Contains(v, " ") {
				return k, strings.Fields(v)
			}

			return k, v
		},
	}), nil); err != nil {
		return &Config{}, fmt.Errorf("config load env overrides: %w", err)
	}

	// 4.: Unmarshal в struct
	var cfg Config
	if err := k.Unmarshal("", &cfg); err != nil {
		return &Config{}, fmt.Errorf("config unmarshal: %w", err)
	}

	// Если ENV_MODE задан, а env-поле не заполнено из файлов/переменных.
	if cfg.Env == "" && mode != "" {
		cfg.Env = mode
	}

	return &cfg, nil
}

func loadYamlFile(k *koanf.Koanf, path string, require bool) error {
	if _, err := os.Stat(path); err != nil {
		if require {
			return fmt.Errorf("config file not found [%s]: %w", path, err)
		}
	}
	if err := k.Load(file.Provider(path), yaml.Parser()); err != nil {
		return fmt.Errorf("config load error [%s]: %w", path, err)
	}
	return nil
}
