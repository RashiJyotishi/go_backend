package config

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func ConnectDB() {
	// connStr := "host=localhost user=rashij password=Rashi*123 dbname=hisabkitab sslmode=disable"
	dbURL := os.Getenv("DATABASE_URL")

    if dbURL == "" {
		log.Fatal("DATABASE_URL not found in environment")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	DB = db
	log.Println("PostgreSQL connected")
}
