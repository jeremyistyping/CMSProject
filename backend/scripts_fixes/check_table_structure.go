package main

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	dsn := "postgres://postgres:postgres@localhost/sistem_akuntans_test?sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Query to get column information
	var columns []struct {
		ColumnName string `gorm:"column:column_name"`
		DataType   string `gorm:"column:data_type"`
	}

	err = db.Raw(`
		SELECT column_name, data_type 
		FROM information_schema.columns 
		WHERE table_name = 'sales'
		ORDER BY ordinal_position
	`).Scan(&columns).Error

	if err != nil {
		log.Fatal("Error querying columns:", err)
	}

	fmt.Println("Sales table structure:")
	fmt.Println("---------------------")
	for _, col := range columns {
		fmt.Printf("  %s (%s)\n", col.ColumnName, col.DataType)
	}
}