package db

import (
	"database/sql"
	"fmt"

	"batch-data-migration/config" // import package config

	_ "github.com/denisenkom/go-mssqldb" // Driver untuk MS SQL Server
)

// ConnectDB untuk menghubungkan ke PostgreSQL atau MS SQL Server menggunakan konfigurasi yang sudah dimuat
func ConnectDB() (*sql.DB, error) {
	// Ambil konfigurasi database dari AppConfig
	dbConfig := config.AppConfig.Database

	var connStr string
	var driverName string

	if dbConfig.Type == "postgres" {
		// Koneksi untuk PostgreSQL
		connStr = fmt.Sprintf("user=%s dbname=%s password=%s host=%s port=%d sslmode=%s",
			dbConfig.User, dbConfig.DBName, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.SSLMode)
		driverName = "postgres"
	} else if dbConfig.Type == "mssql" {
		// Koneksi untuk MS SQL Server
		// Fix: Gunakan server= format yang benar untuk MS SQL Server
		connStr = fmt.Sprintf("server=%s;user id=%s;password=%s;port=%d;database=%s;encrypt=false;TrustServerCertificate=%t",
			dbConfig.Host, dbConfig.User, dbConfig.Password, dbConfig.Port, dbConfig.DBName, dbConfig.TrustServerCertificate)

		if dbConfig.Instance != "" {
			connStr = fmt.Sprintf("%s;instance=%s", connStr, dbConfig.Instance)
		}

		driverName = "sqlserver"

		// Debug connection string (tanpa password)
		fmt.Printf("Connection string: server=%s;user id=%s;port=%d;database=%s\n",
			dbConfig.Host, dbConfig.User, dbConfig.Port, dbConfig.DBName)
	} else {
		return nil, fmt.Errorf("unsupported database type: %s", dbConfig.Type)
	}

	// Membuka koneksi ke database
	db, err := sql.Open(driverName, connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	// Memastikan koneksi ke database berhasil
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	fmt.Printf("Database Host: %s\n", dbConfig.Host)
	fmt.Printf("Database Config: Host=%s, Port=%d, User=%s, DBName=%s, Type=%s\n",
		dbConfig.Host, dbConfig.Port, dbConfig.User, dbConfig.DBName, dbConfig.Type)
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
