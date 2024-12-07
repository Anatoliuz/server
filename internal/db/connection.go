package db

import (
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// InitDB initializes the database connection and performs migrations.
func InitDB() {
	// Retrieve the database connection string from environment variables
	dsn := os.Getenv("DB_DSN") // Example: "host=localhost user=postgres password=postgres dbname=test port=5432 sslmode=disable"
	var err error

	// Open the database connection
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Perform auto-migration for Client and Command models
	err = DB.AutoMigrate(&Client{}, &Command{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	log.Println("Database connected and migrations applied successfully.")
}
