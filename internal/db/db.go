package db

import (
	_ "context"
	"database/sql"
	"fmt"
	_ "fmt"

	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

func Open(path string) (*Store, error) {
	dsn := fmt.Sprintf("file:%s?_journal_mode=WAL&_busy_timeout=5000")
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS page_views (
			id    INTEGER PRIMARY KEY CHECK (id = 1),
			count INTEGER NOT NULL DEFAULT 0
		);
		INSERT OR IGNORE INTO page_views (id, count) VALUES (1, 0);
	`)
	if err != nil {
		return nil, err
	}

	return &Store{
		db: db,
	}, nil

}
