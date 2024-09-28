package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
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

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users(
 		id SERIAL PRIMARY KEY,
 		email TEXT NOT NULL UNIQUE,
		passHash TEXT NOT NULL,
		deleted BOOLEAN NOT NULL DEFAULT false,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP);
		CREATE INDEX IF NOT EXISTS idx_email ON users(email);

		CREATE TABLE IF NOT EXISTS users_data(
     	id SERIAL PRIMARY KEY,
     	user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
     	data_type TEXT NOT NULL,
     	data JSONB NOT NULL,
     	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
     	deleted BOOLEAN DEFAULT false);
	`)
	if err != nil {
		return nil, err
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveUser(ctx context.Context, email string, passHash []byte) (models.User, error) {
	//stmt, err := s.db.Prepare("INSERT INTO users(email, passHash) VALUES ($1, $2) RETURNING id")
	//if err != nil {
	//	return 0, fmt.Errorf("failed to prepare statement: %w", err)
	//}

	stmt, err := s.db.Prepare("INSERT INTO users(email, passHash) VALUES ($1, $2)")
	if err != nil {
		return models.User{}, fmt.Errorf("failed to prepare statement: %w", err)
	}

	_, err = stmt.ExecContext(ctx, email, passHash)
	if err != nil {
		var pgErr *pq.Error

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return models.User{}, storage.ErrUserExists
		}

		return models.User{}, fmt.Errorf("failed to execute statement: %w", err)
	}

	user, err := s.User(ctx, email)
	if err != nil {
		return models.User{}, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
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

func (s *Storage) SaveData(ctx context.Context, userID int64, dataType string, data string) error {
	stmt, err := s.db.Prepare("INSERT INTO users_data(user_id, data_type, data) VALUES ($1, $2, $3)")
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}

	dataJSON, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	var id int64
	err = stmt.QueryRowContext(ctx, userID, dataType, dataJSON).Scan(&id)
	if err != nil {
		return fmt.Errorf("failed to execute statement: %w", err)
	}

	return nil
}

func (s *Storage) GetData(ctx context.Context, userID int64) ([]models.Data, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT data_type, data FROM users_data WHERE user_id = $1", userID)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var allData []models.Data

	for rows.Next() {
		var data models.Data

		err := rows.Scan(&data.DataType, &data.Content)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		allData = append(allData, data)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	if len(allData) == 0 {
		return nil, storage.ErrDataNotFound
	}

	return allData, nil
}
