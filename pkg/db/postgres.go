package repository

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

func InitPostgresDB(connectionString string) (*sql.DB, error) {
  db, err := sql.Open("postgres", connectionString)
  if err != nil {
    log.Fatalf("Failed to connect to database: %v", err)
    return nil, err
  }

  if err := db.Ping(); err != nil {
    log.Fatalf("Failed to ping database: %v", err)
    return nil, err
  }

  log.Println("Successful connected to PostgreSQL")
  return db, nil
}
