package domain

import "github.com/google/uuid"

// UserRepository - интерфейс для работы с пользователями
type UserRepository interface {
	GetUserByChatID(chatID int64) (*User, error)
	CreateUser(user *User) error
	UpdateUser(user *User) error
	UpdateUserLanguage(chatID int64, locale string) error // Новый метод
}

// AuthAttemptRepository - интерфейс для работы с попытками авторизации
type AuthAttemptRepository interface {
	GetPendingAuthAttemptByTgID(tgID int64) (*AuthAttempt, error) // Новый метод
	CreateAuthAttempt(attempt *AuthAttempt) error
	UpdateAuthAttempt(attempt *AuthAttempt) error
	GetExpiredAuthAttempts() ([]AuthAttempt, error)
	GetAuthAttemptByID(id uuid.UUID) (*AuthAttempt, error)
	UpdateAuthAttemptStatus(authID uuid.UUID, statusID int) error
	DeleteExpiredAuthAttempts() error
}

// AuthAttemptStatusRepository - интерфейс для работы со статусами авторизации
type AuthAttemptStatusRepository interface {
	GetStatusByID(id int) (*AuthAttemptStatus, error)
}
