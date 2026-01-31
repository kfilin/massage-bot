package storage

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var DB *sqlx.DB

func InitDB() (*sqlx.DB, error) {
	log.Printf("DEBUG: Connecting to DB: host=%s, port=%s, user=%s, dbname=%s, sslmode=%s",
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"), os.Getenv("DB_NAME"), os.Getenv("DB_SSL_MODE"))
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s connect_timeout=10",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSL_MODE"),
	)

	var db *sqlx.DB
	var err error

	// Attempt to connect to the database.
	// We use a basic retry logic for safety, but primary orchestration
	// should be handled via Docker healthchecks and depends_on.
	for i := 1; i <= 3; i++ {
		db, err = sqlx.Connect("postgres", connStr)
		if err == nil {
			break
		}
		log.Printf("Waiting for database connection (attempt %d/3)...", i)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		return nil, fmt.Errorf("database connection failed: %w. Please ensure the DB container is healthy.", err)
	}

	// Initialize schema
	log.Println("DEBUG: Ensuring database schema is up to date...")
	if _, err := db.Exec(Schema); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}
	log.Println("DEBUG: Database schema initialized/verified.")

	DB = db
	log.Println("Database initialized successfully")
	return db, nil
}
