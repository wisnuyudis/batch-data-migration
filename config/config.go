package config

import (
	"fmt"
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
	Type                   string `yaml:"type"` // postgres atau mssql
	Host                   string `yaml:"host"`
	Port                   int    `yaml:"port"`
	User                   string `yaml:"user"`
	Password               string `yaml:"password"`
	DBName                 string `yaml:"dbname"`
	SSLMode                string `yaml:"sslmode"`                  // Hanya untuk PostgreSQL
	Instance               string `yaml:"instance"`                 // Hanya untuk MS SQL Server
	TrustServerCertificate bool   `yaml:"trust_server_certificate"` // Hanya untuk MS SQL Server
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
	fmt.Printf("Loading configuration from: %s\n", path)
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

	// Tambahkan log untuk memastikan nilai konfigurasi
	fmt.Printf("Loaded Database Config: Host=%s, Port=%d, User=%s, DBName=%s\n",
		AppConfig.Database.Host, AppConfig.Database.Port, AppConfig.Database.User, AppConfig.Database.DBName)
}
