package models

// User - структура для хранения данных о пользователе
type User struct {
	ID       int64
	Email    string
	PassHash []byte
}
