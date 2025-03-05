package db

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

func InitDB() *sql.DB {
	db, err := sql.Open("mysql", "user:password@/dbname")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	return db
}
