package migration

import (
	"batch-data-migration/config"
	"batch-data-migration/internal/db"
	"batch-data-migration/pkg/tokenizer"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/schollz/progressbar/v3"
)

const batchSize = 50

// Initialize logger
var logger *log.Logger

func init() {
	// Membuka log file untuk menyimpan log migrasi
	logFile, err := os.OpenFile("migration.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal("Error opening log file:", err)
	}
	// Membuat logger dengan output ke file log
	logger = log.New(logFile, "MIGRATION: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func RunMigration() {
	// Menyimpan waktu mulai eksekusi
	startTime := time.Now()

	// Log saat migrasi dimulai
	logger.Println("Migration started")

	// Menghubungkan ke database
	conn, err := db.ConnectDB()
	if err != nil {
		logger.Fatal("Database connection failed:", err)
	}
	defer conn.Close()

	// Membuat helper untuk tokenisasi
	helper := tokenizer.NewHelper()

	// Mendapatkan informasi tabel dan kolom dari konfigurasi
	table := config.AppConfig.Migration.Table
	idField := config.AppConfig.Migration.IdField
	targetField := config.AppConfig.Migration.TargetField

	// Menghitung total jumlah baris di tabel untuk progress bar
	var totalCount int
	err = conn.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&totalCount)
	if err != nil {
		logger.Fatal("Error counting rows:", err)
	}

	// Membuat progress bar
	bar := progressbar.NewOptions(totalCount,
		progressbar.OptionSetWidth(50),
		progressbar.OptionSetDescription("Migrating..."),
		progressbar.OptionSetRenderBlankState(true),
	)

	offset := 0
	for {
		// Menarik data sesuai batch
		query := fmt.Sprintf("SELECT %s, %s FROM %s ORDER BY %s LIMIT $1 OFFSET $2", idField, targetField, table, idField)
		rows, err := conn.Query(query, batchSize, offset)
		if err != nil {
			logger.Fatal("Error querying data:", err)
		}

		var count int
		for rows.Next() {
			var id int
			var ccNumber string

			if err := rows.Scan(&id, &ccNumber); err != nil {
				logger.Println("Error scanning row:", err)
				continue
			}

			// Melakukan tokenisasi pada ccNumber
			token, err := helper.Tokenize(ccNumber)
			if err != nil {
				logger.Printf("Error tokenizing cc_number %s: %v\n", ccNumber, err)
				continue
			}

			// Update data dengan token hasil tokenisasi
			updateQuery := fmt.Sprintf("UPDATE %s SET %s = $1 WHERE %s = $2", table, targetField, idField)
			_, err = conn.Exec(updateQuery, token, id)
			if err != nil {
				logger.Printf("Error updating token for id %d: %v\n", id, err)
				continue
			}

			// Menyimpan pesan "Row ID ... updated successfully" hanya ke log
			logMessage := fmt.Sprintf("Row ID %d updated successfully", id)
			logger.Println(logMessage)

			// Menampilkan progress bar
			bar.Add(1)
			count++
		}

		rows.Close()

		// Jika jumlah data yang diambil kurang dari batchSize, migrasi selesai
		if count < batchSize {
			break
		}
		offset += batchSize
	}

	// Menyimpan waktu selesai eksekusi
	endTime := time.Now()

	// Menghitung durasi eksekusi
	duration := endTime.Sub(startTime)

	// Log selesai migrasi
	logger.Printf("Migration complete. Total execution time: %v\n", duration)
	fmt.Printf("\nMigration complete. Total execution time: %v\n", duration)
}
