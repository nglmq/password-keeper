package jwt

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/nglmq/password-keeper/internal/domain/models"
	"time"
)

const superSecret = "superSecret"

func NewToken(user models.User, duration time.Duration) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["uid"] = user.ID
	claims["email"] = user.Email
	claims["exp"] = time.Now().Add(duration).Unix()

	tokenString, err := token.SignedString([]byte(superSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
