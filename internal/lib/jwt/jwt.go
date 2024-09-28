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

func ValidateToken(tokenString string) (int64, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(superSecret), nil
	})
	if err != nil {
		return 0, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, jwt.ErrTokenInvalidClaims
	}

	uid, ok := claims["uid"].(float64)
	if !ok {
		return 0, jwt.ErrInvalidKeyType
	}

	return int64(uid), nil
}
