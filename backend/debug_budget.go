package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:Moon@localhost:5432/CMSNew?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var dataType string
	var precision, scale sql.NullInt64
	err = db.QueryRow(`
		SELECT data_type, numeric_precision, numeric_scale
		FROM information_schema.columns 
		WHERE table_name = 'projects' AND column_name = 'budget'
	`).Scan(&dataType, &precision, &scale)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Budget column type: %s, Precision: %d, Scale: %d\n", dataType, precision.Int64, scale.Int64)
}
