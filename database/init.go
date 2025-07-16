package db

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

func Init(connStr string) (*sql.DB, error) {
	fmt.Println("[DATABASE] ⏳ Opening the database connection...")
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Println("[DATABASE] 🚫 Failed to open database connection")
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	err = db.Ping()
	if err != nil {
		fmt.Println("[DATABASE] 🚫 Failed to ping database")
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	fmt.Println("[DATABASE] ✅ Successfully connected to the database")
	return db, nil
}

func Close(db *sql.DB) error {
	fmt.Println("[DATABASE] ⏳ Closing the database connection...")
	err := db.Close()
	if err != nil {
		fmt.Println("[DATABASE] 🚫 Failed to close database connection")
		return fmt.Errorf("failed to close database connection: %w", err)
	}

	fmt.Println("[DATABASE] ✅ Database connection closed")
	return nil
}
