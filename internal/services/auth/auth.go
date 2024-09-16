package auth

import (
	"context"
	"errors"
	"fmt"
	"github.com/nglmq/password-keeper/internal/lib/jwt"
	"log/slog"
	"time"

	"github.com/nglmq/password-keeper/internal/domain/models"
	"github.com/nglmq/password-keeper/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

type Auth struct {
	log        *slog.Logger
	userSaver  Saver
	userGetter Getter
}

type Saver interface {
	SaveUser(ctx context.Context, email string, passHash []byte) (uid int64, err error)
}

type Getter interface {
	User(ctx context.Context, email string) (models.User, error)
}

// New returns a new instanse of Auth service
func New(log *slog.Logger, userSaver Saver, userGetter Getter) *Auth {
	return &Auth{
		log:        log,
		userSaver:  userSaver,
		userGetter: userGetter,
	}
}

var (
	ErrInvalidCredentials = errors.New("wrong login or password")
)

// Login check credentials and if user exists
func (a *Auth) Login(ctx context.Context, email string, password string) (string, error) {
	log := a.log.With(
		slog.String("method", "Login"),
		slog.String("email", email),
	)

	log.Info("logging in")

	user, err := a.userGetter.User(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Error("user not found", err)

			return "", fmt.Errorf("%w", ErrInvalidCredentials)
		}

		log.Error("failed to get user", err)

		return "", fmt.Errorf("failed to get user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		a.log.Info("invalid credentials", err)

		return "", fmt.Errorf("%w", ErrInvalidCredentials)
	}

	token, err := jwt.NewToken(user, 3*time.Hour)
	if err != nil {
		a.log.Error("failed to generate token", err)

		return "", fmt.Errorf("%w", err)
	}

	return token, nil
}

// Register register new user and returns user ID
func (a *Auth) Register(ctx context.Context, email string, password string) (int64, error) {
	log := a.log.With(
		slog.String("method", "RegisterNewUser"),
		slog.String("email", email),
	)

	log.Info("registering new user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to hash password", err)

		return 0, fmt.Errorf("failed to hash password: %w", err)
	}

	id, err := a.userSaver.SaveUser(ctx, email, passHash)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			log.Error("user already exists", err)

			return 0, fmt.Errorf("user already exists: %w", storage.ErrUserExists)
		}
		log.Error("failed to save user", err)

		return 0, fmt.Errorf("failed to save user: %w", err)
	}

	return id, nil
}
