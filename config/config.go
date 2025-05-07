package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type TokenizationConfig struct {
	BaseURL          string `yaml:"base_url"`
	TokenGroup       string `yaml:"token_group"`
	TokenTemplate    string `yaml:"token_template"`
	TokenizeUser     string `yaml:"tokenize_user"`
	TokenizePassword string `yaml:"tokenize_password"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	SSLMode  string `yaml:"sslmode"`
}

type MigrationConfig struct {
	Table       string `yaml:"table"`
	IdField     string `yaml:"id_field"`
	TargetField string `yaml:"target_field"`
}

type Config struct {
	Tokenization TokenizationConfig `yaml:"tokenization"`
	Database     DatabaseConfig     `yaml:"database"`
	Migration    MigrationConfig    `yaml:"migration"`
}

var AppConfig *Config

func LoadConfig(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}
	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		log.Fatalf("Failed to unmarshal config: %v", err)
	}
	AppConfig = &cfg
}
