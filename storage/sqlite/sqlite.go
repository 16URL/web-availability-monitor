// internal/storage/sqlite/sqlite.go

package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
	"webaviable/storage"

	"github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

var _ storage.Storage = (*Storage)(nil)

func New(storagePath string) (*Storage, error) {
	const op = "storage.sqlite.NewStorage"

	dsn := fmt.Sprintf("%s?_foreign_keys=on", storagePath)
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = db.Exec(`
    CREATE TABLE IF NOT EXISTS url (
        id INTEGER PRIMARY KEY,
        alias TEXT NOT NULL UNIQUE,
        url TEXT NOT NULL
    	);
	`)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS checks(
		id INTEGER PRIMARY KEY,
		url_id INTEGER NOT NULL,
		http_status INTEGER NOT NULL,
		response_time_ms INTEGER NOT NULL,
        checked_at DATETIME NOT NULL,
		FOREIGN KEY(url_id) REFERENCES url(id) ON DELETE CASCADE
	);
	CREATE INDEX IF NOT EXISTS idx_checks_url_checked_at
	ON checks(url_id, checked_at);
	`)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}
func (s *Storage) DeleteURL(alias string) error {
	const op = "storage.sqlite.DeleteURL"

	stmt, err := s.db.Prepare("DELETE FROM url WHERE alias = ?;")
	if err != nil {
		return fmt.Errorf("%s, trouble with delete: %w", op, err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(alias)
	if err != nil {
		return fmt.Errorf("%s, execute statement: %w", op, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s, rows affected: %w", op, err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("%s: %w", op, storage.ErrURLNotFound)
	}
	return nil
}

func (s *Storage) SaveURL(alias string, url string) (int64, error) {
	const op = "storage.sqlite.SaveURL"
	var sqliteErr sqlite3.Error

	stmt, err := s.db.Prepare("INSERT INTO url(url, alias) VALUES(?,?)")
	if err != nil {
		return 0, fmt.Errorf("%s, prepare statement: %w", op, err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(url, alias)
	if err != nil {
		if errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrURLExists)
		}
		return 0, fmt.Errorf("%s, execute statement: %w", op, err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s, get last insert id: %w", op, err)
	}

	return id, nil
}

func (s *Storage) SaveCheck(urlID int64, httpStatus int, responseTimeMs int64, checkedAt time.Time) (int64, error) {
	const op = "storage.sqlite.SaveCheck"

	stmt, err := s.db.Prepare("INSERT INTO checks(url_id, http_status, response_time_ms, checked_at) VALUES(?,?,?,?)")
	if err != nil {
		return 0, fmt.Errorf("%s, prepare statement: %w", op, err)
	}
	defer stmt.Close()

	checkedAt = checkedAt.UTC()

	res, err := stmt.Exec(urlID, httpStatus, responseTimeMs, checkedAt)
	if err != nil {
		return 0, fmt.Errorf("%s, execute statement: %w", op, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s, get last insert id: %w", op, err)
	}

	return id, nil
}

func (s *Storage) GetChecks(urlID int64) ([]storage.Check, error) {
	const op = "storage.sqlite.GetChecks"
	since := time.Now().UTC().Add(-24 * time.Hour)

	rows, err := s.db.Query("SELECT id, url_id, http_status, response_time_ms, checked_at FROM checks WHERE url_id = ? AND checked_at >= ? ORDER BY checked_at DESC", urlID, since)
	if err != nil {
		return nil, fmt.Errorf("%s, query checks: %w", op, err)
	}
	defer rows.Close()

	var checks []storage.Check
	for rows.Next() {
		var check storage.Check
		err := rows.Scan(&check.ID, &check.URLID, &check.HTTPStatus, &check.ResponseTimeMs, &check.CheckedAt)
		if err != nil {
			return nil, fmt.Errorf("%s, scan check: %w", op, err)
		}
		checks = append(checks, check)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s, iterate checks: %w", op, err)
	}
	return checks, nil
}

func (s *Storage) GetURL(alias string) (storage.URL, error) {
	const op = "storage.sqlite.GetURL"
	var url storage.URL
	err := s.db.QueryRow("SELECT id, alias, url FROM url WHERE alias = ?", alias).Scan(&url.ID, &url.Alias, &url.URL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return storage.URL{}, fmt.Errorf("%s: %w", op, storage.ErrURLNotFound)
		}
		return storage.URL{}, fmt.Errorf("%s, query url: %w", op, err)
	}
	return url, nil
}

func (s *Storage) Close() error {
	const op = "storage.sqlite.Close"
	if err := s.db.Close(); err != nil {
		return fmt.Errorf("%s, close db: %w", op, err)
	}
	return nil
}

func (s *Storage) GetAllURLs() ([]storage.URL, error) {
	const op = "storage.sqlite.GetAllURLs"
	rows, err := s.db.Query("SELECT id, alias, url FROM url ORDER BY id ASC")
	if err != nil {
		return nil, fmt.Errorf("%s, query urls: %w", op, err)
	}
	defer rows.Close()

	var urls []storage.URL
	for rows.Next() {
		var url storage.URL
		err := rows.Scan(&url.ID, &url.Alias, &url.URL)
		if err != nil {
			return nil, fmt.Errorf("%s, scan url: %w", op, err)
		}
		urls = append(urls, url)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s, iterate urls: %w", op, err)
	}
	return urls, nil
}
