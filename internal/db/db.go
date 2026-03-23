package db

import (
	"context"
	"database/sql"
	"fmt"
)

type Store struct {
	db *sql.DB
}

func OpenPostgres(connStr string) (*Store, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS page_views (
		id    INTEGER PRIMARY KEY CHECK (id = 1),
		count INTEGER NOT NULL DEFAULT 0
	)`)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(`INSERT INTO page_views (id, count) VALUES (1, 0) ON CONFLICT DO NOTHING`)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS matrix_cells (
		row   INTEGER NOT NULL,
		col   INTEGER NOT NULL,
		value INTEGER NOT NULL DEFAULT 0 CHECK (value IN (0,1)),
		PRIMARY KEY (row, col)
	)`)
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

	_, err = tx.ExecContext(ctx, `UPDATE page_views SET count = count + 1 WHERE id = 1`)
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

const MatrixSize = 3 // The matrix is a 3x3

func (s *Store) InitMatrix(ctx context.Context) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx,
		`INSERT INTO matrix_cells (row, col, value) VALUES ($1, $2, 0) ON CONFLICT DO NOTHING`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for r := 0; r < MatrixSize; r++ {
		for c := 0; c < MatrixSize; c++ {
			if _, err := stmt.ExecContext(ctx, r, c); err != nil {
				return err
			}
		}
	}
	return tx.Commit()
}

func (s *Store) GetMatrix(ctx context.Context) ([][]int, error){
	grid := make([][]int, MatrixSize)
	for i := range grid {
		grid[i] = make([]int, MatrixSize)
	}

	sqlRows, err := s.db.QueryContext(ctx,
	`SELECT row, col, value FROM matrix_cells ORDER by row, col`,
	)
	if err != nil {
		return nil, err
	}
	defer sqlRows.Close()

	for sqlRows.Next() {
		var r, c, v int
		if err := sqlRows.Scan(&r, &c, &v); err != nil {
			return nil, err
		}
		grid[r][c] = v
	}
	return grid, sqlRows.Err()
}

func (s* Store) ToggleCell(ctx context.Context, row int, col int) (int, error){
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx,
		`UPDATE matrix_cells SET value = 1 - value WHERE row = $1 AND col = $2`,
		row, col,
	)
	if err != nil {
		return 0, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	if affected == 0 {
		return 0, fmt.Errorf("cell (%d, %d) does not exist", row, col)
	}
	var value int
	err = tx.QueryRowContext(ctx,
		`SELECT value FROM matrix_cells WHERE row = $1 AND col = $2`,
		row, col,
	).Scan(&value)
	if err != nil {
		return 0, err
	}
	return value, tx.Commit()
}
