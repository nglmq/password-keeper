package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jackc/pgerrcode"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/lib/pq"
	"github.com/nglmq/password-keeper/internal/domain/models"
	"github.com/nglmq/password-keeper/internal/storage"
)

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	db, err := sql.Open("pgx", storagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open connection: %w", err)
	}

	stmt, err := db.Prepare(`
		CREATE TABLE IF NOT EXISTS users(
 		id SERIAL PRIMARY KEY,
 		email TEXT NOT NULL UNIQUE,
		passHash TEXT NOT NULL,
		deleted BOOLEAN NOT NULL DEFAULT false,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP);
	`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	_, err = stmt.Exec()
	if err != nil {
		return nil, err
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveUser(ctx context.Context, email string, passHash []byte) (int64, error) {
	//stmt, err := s.db.Prepare("INSERT INTO users(email, passHash) VALUES ($1, $2) RETURNING id")
	//if err != nil {
	//	return 0, fmt.Errorf("failed to prepare statement: %w", err)
	//}

	var id int64

	err := s.db.QueryRowContext(ctx, "INSERT INTO users(email, passHash) VALUES ($1, $2) RETURNING id", email, passHash).Scan(&id)
	if err != nil {
		var pgErr *pq.Error

		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return 0, fmt.Errorf("user already exists: %w", storage.ErrUserExists)
		}

		return 0, fmt.Errorf("failed to execute statement: %w", err)
	}

	return id, nil
}

func (s *Storage) User(ctx context.Context, email string) (models.User, error) {
	stmt, err := s.db.Prepare("SELECT id, email, passHash FROM users WHERE email = $1")
	if err != nil {
		return models.User{}, fmt.Errorf("failed to prepare statement: %w", err)
	}

	row := stmt.QueryRowContext(ctx, email)

	var user models.User

	err = row.Scan(&user.ID, &user.Email, &user.PassHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, fmt.Errorf("user not found: %w", storage.ErrUserNotFound)
		}

		return models.User{}, fmt.Errorf("failed to execute statement: %w", err)
	}

	return user, nil
}
