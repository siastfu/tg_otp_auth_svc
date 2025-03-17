package domain

import "time"

// User - сущность пользователя (таблица user.profile)
type User struct {
	ID          int       `json:"id"`
	Username    string    `json:"username"`
	ChatID      int64     `json:"chat_id"` // Telegram ID
	FullName    string    `json:"full_name"`
	PhoneNumber string    `json:"phone_number"`
	Locale      string    `json:"locale"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// AuthAttempt - попытка авторизации (таблица auth.attempt)
type AuthAttempt struct {
	ID          string     `json:"id"`                     // UUID
	TgID        int64      `json:"tg_id"`                  // Telegram ID пользователя
	StatusID    int        `json:"status_id"`              // ID статуса
	CreatedAt   time.Time  `json:"created_at"`             // Время создания
	ExpiredAt   time.Time  `json:"expired_at"`             // Время истечения
	SucceededAt *time.Time `json:"succeeded_at,omitempty"` // Время успешной авторизации
}

// AuthAttemptStatus - статусы авторизации (таблица list.auth_attempt_status)
type AuthAttemptStatus struct {
	ID           int    `json:"id"`            // ID статуса
	TechName     string `json:"tech_name"`     // Техническое название статуса
	DisplayName  string `json:"display_name"`  // Локализованное название (JSONB)
	DisplayImage string `json:"display_image"` // Изображение статуса (JSONB)
}
