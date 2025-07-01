package migration

import (
	"batch-data-migration/config"
	"batch-data-migration/internal/db"
	"batch-data-migration/pkg/tokenizer"
	"database/sql"
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
		var query string
		var rows *sql.Rows
		var err error

		if config.AppConfig.Database.Type == "postgres" {
			// PostgreSQL syntax with LIMIT/OFFSET
			query = fmt.Sprintf("SELECT %s, %s FROM %s ORDER BY %s LIMIT $1 OFFSET $2", idField, targetField, table, idField)
			rows, err = conn.Query(query, batchSize, offset)
		} else {
			// SQL Server syntax with OFFSET/FETCH
			query = fmt.Sprintf("SELECT %s, %s FROM %s ORDER BY %s OFFSET %d ROWS FETCH NEXT %d ROWS ONLY",
				idField, targetField, table, idField, offset, batchSize)
			rows, err = conn.Query(query)
		}

		if err != nil {
			logger.Fatal("Error querying data:", err)
		}

		type DataRow struct {
			ID       int
			CCNumber string
		}

		var dataRows []DataRow
		for rows.Next() {
			var id int
			var ccNumber string

			if err := rows.Scan(&id, &ccNumber); err != nil {
				rows.Close()
				logger.Fatal("Error scanning row:", err)
			}

			dataRows = append(dataRows, DataRow{ID: id, CCNumber: ccNumber})
		}
		rows.Close()

		// If no rows returned, we're done
		if len(dataRows) == 0 {
			break
		}

		// Extract credit card numbers for batch tokenization
		var ccNumbers []string
		for _, row := range dataRows {
			ccNumbers = append(ccNumbers, row.CCNumber)
		}

		// Tokenize in batch
		tokens, err := helper.TokenizeBatch(ccNumbers)
		if err != nil {
			logger.Fatal("Error tokenizing batch:", err)
		}

		// Begin a transaction for batch update
		tx, err := conn.Begin()
		if err != nil {
			logger.Fatal("Error starting transaction:", err)
		}

		// Prepare the update statement
		var updateStmt *sql.Stmt
		if config.AppConfig.Database.Type == "postgres" {
			updateStmt, err = tx.Prepare(fmt.Sprintf("UPDATE %s SET %s = $1 WHERE %s = $2", table, targetField, idField))
		} else {
			// For SQL Server, use @ parameters instead of ?
			updateStmt, err = tx.Prepare(fmt.Sprintf("UPDATE %s SET %s = @p1 WHERE %s = @p2", table, targetField, idField))
		}

		if err != nil {
			tx.Rollback()
			logger.Fatal("Error preparing update statement:", err)
		}

		// Update each row with its token
		for i, row := range dataRows {
			if i < len(tokens) {
				_, err := updateStmt.Exec(tokens[i], row.ID)
				if err != nil {
					tx.Rollback()
					logger.Fatal("Error updating row:", err)
				}

				// Log update success
				logger.Printf("Row ID %d updated successfully", row.ID)

				// Update progress bar
				bar.Add(1)
			}
		}

		// Commit the transaction
		if err := tx.Commit(); err != nil {
			logger.Fatal("Error committing transaction:", err)
		}

		// Move to next batch
		offset += len(dataRows)
	}

	// Menyimpan waktu selesai eksekusi
	endTime := time.Now()

	// Menghitung durasi eksekusi
	duration := endTime.Sub(startTime)

	// Log selesai migrasi
	logger.Printf("Migration complete. Total execution time: %v\n", duration)
	fmt.Printf("\nMigration complete. Total execution time: %v\n", duration)
}
