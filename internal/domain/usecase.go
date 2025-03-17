package domain

// UserUseCase - интерфейс для работы с пользователями
type UserUseCase interface {
	HandleUserEntry(chatID int64, username, fullName, phoneNumber string) (*User, error)
	UpdateUserLanguage(chatID int64, locale string) error
	GetUserByTgID(tgID int64) (*User, error)
	CreateUser(user *User) error
}

// AuthUseCase - интерфейс бизнес-логики авторизации
type AuthUseCase interface {
	CheckAuthLink(uuid string) (*AuthAttempt, error)
	StartAuthorization(tgID int64) (string, error)
	UpdateAuthAttempt(attempt *AuthAttempt) error
	MarkExpiredAuthAttempts() error
	UpdateUserLanguage(tgID int64, lang string) error // Добавили этот метод
}
