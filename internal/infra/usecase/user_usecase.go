package usecase

import (
	"time"

	"tg_otp_auth_svc/internal/domain"
)

// UserUseCaseImpl - реализация UserUseCase
type UserUseCaseImpl struct {
	UserRepo domain.UserRepository
}

// Убедимся, что UserUseCaseImpl реализует интерфейс UserUseCase
var _ domain.UserUseCase = (*UserUseCaseImpl)(nil)

// NewUserUseCase создает новый экземпляр UserUseCase
func NewUserUseCase(userRepo domain.UserRepository) domain.UserUseCase {
	return &UserUseCaseImpl{UserRepo: userRepo}
}

import "go.uber.org/zap"

func (u *UserUseCaseImpl) HandleUserEntry(chatID int64, username, fullName, phoneNumber string) (*domain.User, error) {
	existingUser, err := u.UserRepo.GetUserByChatID(chatID)
	if err != nil {
		logger.Logger.Error("Ошибка при получении пользователя", zap.Int64("chat_id", chatID), zap.Error(err))
		return nil, err
	}

	now := time.Now()

	if existingUser != nil {
		existingUser.Username = username
		existingUser.FullName = fullName
		existingUser.PhoneNumber = phoneNumber
		existingUser.UpdatedAt = now

		err = u.UserRepo.UpdateUser(existingUser)
		if err != nil {
			logger.Logger.Error("Ошибка при обновлении пользователя", zap.Int64("chat_id", chatID), zap.Error(err))
			return nil, err
		}

		return existingUser, nil
	}

	newUser := &domain.User{
		ChatID:      chatID,
		Username:    username,
		FullName:    fullName,
		PhoneNumber: phoneNumber,
		Locale:      "",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	err = u.UserRepo.CreateUser(newUser)
	if err != nil {
		logger.Logger.Error("Ошибка при создании пользователя", zap.Int64("chat_id", chatID), zap.Error(err))
		return nil, err
	}

	logger.Logger.Info("Пользователь успешно зарегистрирован", zap.Int64("chat_id", chatID))
	return newUser, nil
}

// UpdateUserLanguage обновляет язык пользователя
func (u *UserUseCaseImpl) UpdateUserLanguage(chatID int64, locale string) error {
	user, err := u.UserRepo.GetUserByChatID(chatID)
	if err != nil {
		return err
	}

	if user == nil {
		return nil // Если пользователя нет, просто ничего не делаем
	}

	user.Locale = locale
	user.UpdatedAt = time.Now()

	return u.UserRepo.UpdateUser(user)
}
