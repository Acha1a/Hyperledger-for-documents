package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func Connect() {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "admin"),
		getEnv("DB_PASSWORD", "ToJo"),
		getEnv("DB_NAME", "documents_db"),
	)

	var err error	
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to open DB:", err)
	}

	// Проверка подключения
	err = DB.Ping()
	if err != nil {
		log.Fatal("Failed to connect to DB:", err)
	}

	log.Println("Connected to PostgreSQL")
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}