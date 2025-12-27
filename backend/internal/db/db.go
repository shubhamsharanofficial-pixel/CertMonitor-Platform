package db

import (
	"database/sql"
	_ "embed" // Essential for the go:embed directive
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

//go:embed schema.sql
var schemaSQL string

type Store struct {
	Conn *sql.DB
}

func NewPostgresStore(connStr string) (*Store, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	log.Println("✅ Connected to PostgreSQL")
	return &Store{Conn: db}, nil
}

// InitSchema executes the SQL from the embedded file
func (s *Store) InitSchema() error {
	_, err := s.Conn.Exec(schemaSQL)
	if err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}

	log.Println("✅ Database schema initialized successfully")
	return nil
}
