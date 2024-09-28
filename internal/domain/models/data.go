package models

type Data struct {
	DataType string //`json:"data_type"` // Тип данных (например, текст, пароль, карта и т.д.)
	Content  string //`json:"content"`   // Основное содержимое данных
}
