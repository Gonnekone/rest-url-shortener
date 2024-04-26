package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/Gonnekone/rest-url-shortener/internal/storage"
	"github.com/redis/go-redis/v9"
)

type RedisStorage struct {
	db *redis.Client
}

func (s *RedisStorage) Close() error {
	return s.db.Close()
}

func New() (*RedisStorage, error) {
	const op = "storage.sqlite.New"

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
		MaxRetries: 3,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &RedisStorage{db: rdb}, nil
}

func (s *RedisStorage) SaveURL(urlToSave string, alias string) error {
	const op = "storage.redis.SaveURL"

	fmt.Println("saving url in Redis")

	err := s.db.Set(context.Background(), alias, urlToSave, 10*time.Second).Err()
    if err != nil {
		return fmt.Errorf("%s: %w", op, err)
    }

	return nil
}

func (s *RedisStorage) GetURL(alias string) (string, error) {
	const op = "storage.redis.GetURL"

	fmt.Println("getting url from Redis")

	val, err := s.db.Get(context.Background(), alias).Result()
    if err != nil {
		if err == redis.Nil {
			return "", storage.ErrURLNotFound
		}
		return "", fmt.Errorf("%s: %w", op, err)
    }

	fmt.Println("used redis to pull url"+val)

	return val, nil
}