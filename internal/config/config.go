package config

import (
	"os"

	yaml "gopkg.in/yaml.v3"
)

type Config struct {
	Port                string   `yaml:"port" default:"8000"`
	ValidTiers          []string `yaml:"valid_tiers"`
	S3Bucket            string   `yaml:"s3_bucket"`
	BatchSizeMB         int      `yaml:"batch_size_mb"`
	BatchTimeoutSeconds int      `yaml:"batch_timeout_seconds"`
	AWSRegion           string   `yaml:"aws_region"`
}

func LoadConfig() (*Config, error) {
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if cfg.S3Bucket == "" {
		cfg.S3Bucket = "your-name_portal26_events"
	}
	if cfg.BatchSizeMB == 0 {
		cfg.BatchSizeMB = 5
	}
	if cfg.BatchTimeoutSeconds == 0 {
		cfg.BatchTimeoutSeconds = 5
	}
	if cfg.AWSRegion == "" {
		cfg.AWSRegion = "us-east-1"
	}
	if len(cfg.ValidTiers) == 0 {
		cfg.ValidTiers = []string{"free", "pro", "enterprise"}
	}

	return &cfg, nil
}
