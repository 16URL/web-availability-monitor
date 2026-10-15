// internal/storage/sqlite/sqlite.go

package sqlite

import (
	"database/sql"
	"fmt"
	"time"
)

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	const op = "storage.sqlite.NewStorage"

	db, err := sql.Open("sqlite3", storagePath)
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
		FOREIGN KEY(url_id) REFERENCES url(id)
	);
	`)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveURL(alias string, url string) (int64, error) {
	const op = "storage.sqlite.SaveURL"

	stmt, err := s.db.Prepare("INSERT INTO url(url, alias) VALUES(?,?)")
	if err != nil {
		return 0, fmt.Errorf("%s, prepare statement: %w", op, err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(url, alias)
	if err != nil {
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
