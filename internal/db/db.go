package db

import (
	"database/sql"
	"fmt"

	"batch-data-migration/config" // import package config

	_ "github.com/lib/pq"
)

// ConnectDB untuk menghubungkan ke PostgreSQL menggunakan konfigurasi yang sudah dimuat
func ConnectDB() (*sql.DB, error) {
	// Ambil konfigurasi database dari AppConfig
	dbConfig := config.AppConfig.Database
	connStr := fmt.Sprintf("user=%s dbname=%s password=%s host=%s port=%d sslmode=%s",
		dbConfig.User, dbConfig.DBName, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.SSLMode)

	// Membuka koneksi ke database PostgreSQL
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	// Memastikan koneksi ke database berhasil
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	fmt.Println("Connected to the database successfully!")
	return db, nil
}

// FetchDataBatch untuk mengambil data batch dari database menggunakan offset dan limit
func FetchDataBatch(db *sql.DB, offset, limit int) ([]string, error) {
	// Ambil nama tabel dan kolom target dari konfigurasi migrasi
	table := config.AppConfig.Migration.Table
	targetField := config.AppConfig.Migration.TargetField

	// Mengambil data berdasarkan batch dengan query yang lebih dinamis
	query := fmt.Sprintf("SELECT %s FROM %s LIMIT $1 OFFSET $2", targetField, table)
	rows, err := db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data batch: %v", err)
	}
	defer rows.Close()

	var data []string
	for rows.Next() {
		var ccNumber string
		if err := rows.Scan(&ccNumber); err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}
		data = append(data, ccNumber)
	}

	// Memeriksa apakah ada kesalahan saat iterasi rows
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed during row iteration: %v", err)
	}

	return data, nil
}
