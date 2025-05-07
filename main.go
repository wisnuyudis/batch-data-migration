package main

import (
	"batch-data-migration/config"
	"batch-data-migration/internal/migration"
	"log"
)

func main() {
	// Memuat konfigurasi dari file YAML
	config.LoadConfig("config/config.yaml")
	log.Println("Config loaded:", config.AppConfig)

	// Menjalankan migrasi setelah konfigurasi dimuat
	log.Println("Starting data migration...")
	migration.RunMigration()

	log.Println("Migration completed.")
}
