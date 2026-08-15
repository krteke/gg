package internal

import (
	"os"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	Bucket BucketConfig `toml:"bucket"`
}

type BucketConfig struct {
	Max int64 `toml:"max"`
}

func DefaultConfig() *Config {

	return &Config{
		Bucket: BucketConfig{
			Max: 1024 * 1024 * 1024 * 400,
		},
	}
}

func LoadConfig(path string) (*Config, error) {
	path = strings.TrimSpace(path)
	cfg := DefaultConfig()

	if path == "" {
		return cfg, nil
	}

	file, err := os.Open(path)
	if err != nil {
		error := Err(ErrOpenConfigFile, "open config "+path, err)

		return nil, error
	}
	defer file.Close()

	if err := toml.NewDecoder(file).Decode(cfg); err != nil {
		error := Err(ErrDecodeConfig, "decode config", err)

		return nil, error
	}

	return cfg, nil
}

type ScanConfig struct {
	Root       string
	Output     string
	ConfigPath string
}
