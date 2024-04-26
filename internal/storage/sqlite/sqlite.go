package sqlite

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/Gonnekone/rest-url-shortener/internal/storage"
	"github.com/Gonnekone/rest-url-shortener/internal/storage/redis"
	"github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
	cache *redis.RedisStorage
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func New(storagePath string) (*Storage, error) {
	const op = "storage.sqlite.New"

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w, op, err")
	}

	stmt, err := db.Prepare(`
	CREATE TABLE IF NOT EXISTS url(
		id INTEGER PRIMARY KEY,
		alias TEXT NOT NULL UNIQUE,
		url TEXT NOT NULL);
	CREATE INDEX IF NOT EXISTS idx_alias ON url(alias);
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	rdb, err := redis.New()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	fmt.Println("redis is ready")

	return &Storage{db: db, cache: rdb}, nil
}

func (s *Storage) SaveURL(urlToSave string, alias string) error {
	const op = "storage.sqlite.SaveURL"

	stmt, err := s.db.Prepare("INSERT INTO url(url, alias) VALUES(?, ?)")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.Exec(urlToSave, alias)
	if err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return fmt.Errorf("%s: %w", op, storage.ErrURLExists)
		}

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) GetURL(alias string) (string, error) {
	const op = "storage.sqlite.GetURL"

	resURL, redisErr := s.cache.GetURL(alias)
	if redisErr == nil {
		return resURL, nil
	}

	if redisErr != storage.ErrURLNotFound {
		return "", fmt.Errorf("%s: %w", op, redisErr)
	}

	stmt, err := s.db.Prepare("SELECT url FROM url WHERE alias = ?")
	if err != nil {
		return "", fmt.Errorf("%s: prepare statement: %w", op, err)
	}

	err = stmt.QueryRow(alias).Scan(&resURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", storage.ErrURLNotFound
		}

		return "", fmt.Errorf("%s: execute statement: %w", op, err)
	}

	if errors.Is(redisErr, storage.ErrURLNotFound) {
		s.cache.SaveURL(resURL, alias)
	}

	return resURL, nil
}

func (s *Storage) GetRandomAlias() (string, error) {
    const op = "storage.sqlite.GetRandomAlias"

    query := "SELECT alias FROM url ORDER BY RANDOM() LIMIT 1"

    var alias string
    err := s.db.QueryRow(query).Scan(&alias)
    if err != nil {
        return "", fmt.Errorf("%s: %w", op, err)
    }

	fmt.Println("got random alias")

    return alias, nil
}

func (s* Storage) DeleteURL(alias string) error {
	const op = "storage.sqlite.DeleteURL"

	stmt, err := s.db.Prepare("DELETE FROM url WHERE alias = ?")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.Exec(alias)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
