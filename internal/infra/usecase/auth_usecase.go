package usecase

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"tg_otp_auth_svc/internal/domain"
	"tg_otp_auth_svc/pkg/logger"
)

// AuthUseCaseImpl - реализация AuthUseCase
type AuthUseCaseImpl struct {
	AuthRepo domain.AuthAttemptRepository
	UserRepo domain.UserRepository // ✅ Добавляем UserRepo
}

// Убедимся, что AuthUseCaseImpl реализует интерфейс
var _ domain.AuthUseCase = (*AuthUseCaseImpl)(nil)

// NewAuthUseCase создает новый экземпляр AuthUseCase
func NewAuthUseCase(authRepo domain.AuthAttemptRepository, userRepo domain.UserRepository) domain.AuthUseCase {
	return &AuthUseCaseImpl{
		AuthRepo: authRepo,
		UserRepo: userRepo, // ✅ Передаём UserRepo в конструктор
	}
}

import "go.uber.org/zap"

func (u *AuthUseCaseImpl) CheckAuthLink(id uuid.UUID) (*domain.AuthAttempt, error) {
	authAttempt, err := u.AuthRepo.GetAuthAttemptByID(id)
	if err != nil {
		return nil, err
	}

	if authAttempt == nil {
		return nil, domain.ErrAuthAttemptNotFound
	}

	if authAttempt.StatusID != 1 {
		return nil, domain.ErrAuthAttemptNotPending
	}

	if time.Now().After(authAttempt.ExpiredAt) {
		return nil, domain.ErrAuthLinkExpired
	}

	// Обновляем статус авторизации в БД
	authAttempt.StatusID = 3 // SUCCESS
	now := time.Now()
	authAttempt.SucceededAt = &now

	err = u.AuthRepo.UpdateAuthAttempt(authAttempt)
	if err != nil {
		logger.Logger.Error("Ошибка при обновлении статуса авторизации", zap.Error(err))
		return nil, domain.ErrInternalServerError
	}

	// ✅ Логируем успешную авторизацию
	logger.Logger.Info("Авторизация успешна",
		zap.String("uuid", id.String()),
		zap.Time("succeeded_at", now))

	return authAttempt, nil
}

// UpdateAuthAttempt обновляет статус авторизационной попытки
func (u *AuthUseCaseImpl) UpdateAuthAttempt(authAttempt *domain.AuthAttempt) error {
	return u.AuthRepo.UpdateAuthAttempt(authAttempt)
}

// StartAuthorization начинает процесс авторизации
func (u *AuthUseCaseImpl) StartAuthorization(tgID int64) (string, error) {
	authAttempt, err := u.AuthRepo.GetPendingAuthAttemptByTgID(tgID)
	if err != nil {
		return "", err
	}

	if authAttempt != nil {
		return authAttempt.ID, nil
	}

	newUUID := uuid.New().String()

	newAuthAttempt := &domain.AuthAttempt{
		ID:        newUUID,
		TgID:      tgID,
		StatusID:  1,
		CreatedAt: time.Now(),
		ExpiredAt: time.Now().Add(10 * time.Minute),
	}

	err = u.AuthRepo.CreateAuthAttempt(newAuthAttempt)
	if err != nil {
		return "", err
	}

	return newUUID, nil
}

// MarkExpiredAuthAttempts помечает истекшие попытки авторизации как TIMEOUT
func (u *AuthUseCaseImpl) MarkExpiredAuthAttempts() error {
	authAttempts, err := u.AuthRepo.GetExpiredAuthAttempts()
	if err != nil {
		logger.Logger.Error("Ошибка при получении истекших попыток авторизации", zap.Error(err))
		return err
	}

	if len(authAttempts) == 0 {
		logger.Logger.Info("Нет истекших попыток авторизации для обновления")
		return nil
	}

	logger.Logger.Info("Обновление статусов для истекших попыток авторизации", zap.Int("количество", len(authAttempts)))

	var updateErrors []error

	for _, attempt := range authAttempts {
		if err := u.AuthRepo.UpdateAuthAttemptStatus(attempt.ID, 2); err != nil {
			logger.Logger.Error("Ошибка при обновлении статуса попытки авторизации",
				zap.String("auth_attempt_id", attempt.ID), zap.Error(err))
			updateErrors = append(updateErrors, err)
		}
	}

	if len(updateErrors) > 0 {
		logger.Logger.Error("Ошибки при обновлении статусов попыток авторизации", zap.Int("количество", len(updateErrors)))
		return fmt.Errorf("ошибки обновления статусов: %d", len(updateErrors))
	}

	logger.Logger.Info("Успешно обновлены все истекшие попытки авторизации")
	return nil
}

// ✅ Используем UserRepo вместо AuthRepo
// GetUserByTgID - получает пользователя по Telegram ID
func (u *AuthUseCaseImpl) GetUserByTgID(tgID int64) (*domain.User, error) {
	return u.UserRepo.GetUserByChatID(tgID) // ✅ Меняем на UserRepo
}

// CreateUser - создает нового пользователя
func (u *AuthUseCaseImpl) CreateUser(user *domain.User) error {
	return u.UserRepo.CreateUser(user) // ✅ Меняем на UserRepo
}

// UpdateUserLanguage - обновляет язык пользователя
func (u *AuthUseCaseImpl) UpdateUserLanguage(tgID int64, lang string) error {
	if lang != "ru" && lang != "en" {
		return errors.New("неподдерживаемый язык")
	}
	return u.UserRepo.UpdateUserLanguage(tgID, lang) // ✅ Меняем на UserRepo
}
