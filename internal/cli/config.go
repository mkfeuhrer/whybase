package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ADRDir string `yaml:"adr_dir"`
}

func LoadConfig(root string) (Config, error) {
	cfg := Config{ADRDir: "doc/adr"}
	b, err := os.ReadFile(filepath.Join(root, ".whybase.yml"))
	if os.IsNotExist(err) {
		return cfg, nil
	}
	if err != nil {
		return cfg, err
	}
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return cfg, fmt.Errorf(".whybase.yml: %w", err)
	}
	if cfg.ADRDir == "" {
		cfg.ADRDir = "doc/adr"
	}
	return cfg, nil
}
