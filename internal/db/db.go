package db

import (
	"context"
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
	dsn := fmt.Sprintf("file:%s?_journal_mode=WAL&_busy_timeout=5000", path)
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
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS matrix_cells (
			row   INTEGER NOT NULL,
			col   INTEGER NOT NULL,
			value INTEGER NOT NULL DEFAULT 0 CHECK (value IN (0,1))
			PRIMARY KEY (row, col)
		);
	`)
	if err != nil {
		return nil, err
	}

	return &Store{
		db: db,
	}, nil

}

func (s *Store) GetCount(ctx context.Context) (int64, error){
	var count int64 
	err := s.db.QueryRowContext(ctx, `SELECT count FROM page_views WHERE id = 1`).Scan(&count)
	return count, err
}

func (s *Store) Increment(ctx context.Context) (int64, error){
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `UPDATE page_view s SET count = count + 1 WHERE id = 1`)
	if err != nil {
		return 0, err
	}

	var count int64 
	err = tx.QueryRowContext(ctx, `SELECT count FROM page_views WHERE id = 1`).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, tx.Commit()
}
