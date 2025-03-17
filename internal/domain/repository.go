package domain

// UserRepository - интерфейс для работы с пользователями
type UserRepository interface {
	GetUserByChatID(chatID int64) (*User, error)
	CreateUser(user *User) error
	UpdateUser(user *User) error
	UpdateUserLanguage(chatID int64, locale string) error // Новый метод
}

// AuthAttemptRepository - интерфейс для работы с попытками авторизации
type AuthAttemptRepository interface {
	GetAuthAttemptByID(id string) (*AuthAttempt, error)
	GetPendingAuthAttemptByTgID(tgID int64) (*AuthAttempt, error) // Новый метод
	CreateAuthAttempt(attempt *AuthAttempt) error
	UpdateAuthAttempt(attempt *AuthAttempt) error
	GetExpiredAuthAttempts() ([]AuthAttempt, error)
	UpdateAuthAttemptStatus(authID string, statusID int) error
}

// AuthAttemptStatusRepository - интерфейс для работы со статусами авторизации
type AuthAttemptStatusRepository interface {
	GetStatusByID(id int) (*AuthAttemptStatus, error)
}
