package auth

import (
	"context"
	"log/slog"

	"github.com/nglmq/password-keeper/internal/domain/models"
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

// Login check credentials and if user exists
func (a *Auth) Login(ctx context.Context, email string, password string) (string, error) {

}

// RegisterNewUser register new user and returns user ID
func (a *Auth) RegisterNewUser(ctx context.Context, email string, password string) (int64, error) {

}
