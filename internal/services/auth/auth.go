package auth

import (
	"context"
	"errors"
	"fmt"
	"github.com/nglmq/password-keeper/internal/lib/crypt"
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

type Data struct {
	log        *slog.Logger
	dataSaver  DataSaver
	dataGetter DataGetter
}

type Saver interface {
	SaveUser(ctx context.Context, email string, passHash []byte) (models.User, error)
}

type Getter interface {
	User(ctx context.Context, email string) (models.User, error)
}

type DataSaver interface {
	SaveData(ctx context.Context, userID int64, dataType, data string) error
}

type DataGetter interface {
	GetData(ctx context.Context, userID int64) ([]models.Data, error)
}

// NewAuth returns a new instanse of Auth service
func NewAuth(log *slog.Logger, userSaver Saver, userGetter Getter) *Auth {
	return &Auth{
		log:        log,
		userSaver:  userSaver,
		userGetter: userGetter,
	}
}

func NewData(log *slog.Logger, dataSaver DataSaver, dataGetter DataGetter) *Data {
	return &Data{
		log:        log,
		dataSaver:  dataSaver,
		dataGetter: dataGetter,
	}
}

var (
	ErrInvalidCredentials = errors.New("wrong login or password")
	ErrUserAlreadyExists  = errors.New("user already exists")
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

			return "", ErrInvalidCredentials
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

// RegisterNewUser register new user and returns user ID
func (a *Auth) RegisterNewUser(ctx context.Context, email string, password string) (string, error) {
	log := a.log.With(
		slog.String("method", "RegisterNewUser"),
		slog.String("email", email),
	)

	log.Info("registering new user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to hash password", err)

		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	user, err := a.userSaver.SaveUser(ctx, email, passHash)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			log.Error("user already exists", err)

			return "", ErrUserAlreadyExists
		}
		log.Error("failed to save user", err)

		return "", fmt.Errorf("failed to save user: %w", err)
	}

	token, err := jwt.NewToken(user, 3*time.Hour)
	if err != nil {
		a.log.Error("failed to generate token", err)

		return "", fmt.Errorf("%w", err)
	}

	return token, nil
}

func (d *Data) SaveData(ctx context.Context, token, dataType, data string) (string, error) {
	log := d.log.With(
		slog.String("method", "SaveData"),
		slog.String("dataType", dataType),
	)

	userID, err := jwt.ValidateToken(token)
	if err != nil {
		log.Error("failed to validate token", err)

		return "", fmt.Errorf("failed to validate token: %w", err)
	}

	cr, err := crypt.NewCrypt()
	if err != nil {
		log.Error("failed to create crypt", err)
		return "", fmt.Errorf("failed to create crypt: %w", err)
	}

	// Шифруем данные перед сохранением
	encryptedData := cr.Encode(data)

	log.Info("saving data")

	err = d.dataSaver.SaveData(ctx, userID, dataType, encryptedData)
	if err != nil {
		log.Error("failed to save data", err)

		return "", fmt.Errorf("failed to save data: %w", err)
	}

	return token, nil
}

func (d *Data) GetData(ctx context.Context, token string) (string, []models.Data, error) {
	log := d.log.With(
		slog.String("method", "GetData"),
		slog.String("token", token),
	)

	userID, err := jwt.ValidateToken(token)
	if err != nil {
		log.Error("failed to validate token", err)

		return "", []models.Data{}, fmt.Errorf("failed to validate token: %w", err)
	}

	log.Info("getting data")

	data, err := d.dataGetter.GetData(ctx, userID)
	if err != nil {
		log.Error("failed to get data", err)
		if errors.Is(err, storage.ErrDataNotFound) {
			return token, []models.Data{}, nil
		}

		return token, []models.Data{}, err
	}

	cr, err := crypt.NewCrypt()
	if err != nil {
		log.Error("failed to create crypt", err)
		return token, []models.Data{}, fmt.Errorf("failed to create crypt: %w", err)
	}

	for i := range data {
		data[i].Content, err = cr.Decode(data[i].Content)
		if err != nil {
			log.Error("failed to decode data", err)
			return token, []models.Data{}, fmt.Errorf("failed to decode data: %w", err)
		}
	}

	return token, data, nil
}
