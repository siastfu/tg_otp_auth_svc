package usecase

import (
	"fmt"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"tg_otp_auth_svc/internal/domain"
	"tg_otp_auth_svc/pkg/logger"
	"time"
)

// AuthUseCaseImpl - реализация AuthUseCase
type AuthUseCaseImpl struct {
	AuthRepo domain.AuthAttemptRepository
}
type AuthUseCase interface {
	MarkExpiredAuthAttempts() error
}

// Убедимся, что AuthUseCaseImpl реализует интерфейс
var _ domain.AuthUseCase = (*AuthUseCaseImpl)(nil)

// NewAuthUseCase создает новый экземпляр AuthUseCase
func NewAuthUseCase(authRepo domain.AuthAttemptRepository) domain.AuthUseCase {
	return &AuthUseCaseImpl{AuthRepo: authRepo}
}

// CheckAuthLink проверяет статус авторизационной ссылки
func (u *AuthUseCaseImpl) CheckAuthLink(uuid string) (*domain.AuthAttempt, error) {
	authAttempt, err := u.AuthRepo.GetAuthAttemptByID(uuid)
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

func (u *AuthUseCaseImpl) MarkExpiredAuthAttempts() error {
	// Получаем все истекшие попытки авторизации
	authAttempts, err := u.AuthRepo.GetExpiredAuthAttempts()
	if err != nil {
		logger.Logger.Error("Ошибка при получении истекших попыток авторизации", zap.Error(err))
		return err
	}

	if len(authAttempts) == 0 {
		logger.Logger.Info("Нет истекших попыток авторизации для обновления")
		return nil
	}

	// Логируем количество найденных записей
	logger.Logger.Info("Обновление статусов для истекших попыток авторизации", zap.Int("количество", len(authAttempts)))

	var updateErrors []error // Список ошибок при обновлении

	// Обновляем статус каждой попытки
	for _, attempt := range authAttempts {
		if err := u.AuthRepo.UpdateAuthAttemptStatus(attempt.ID, 2); err != nil {
			logger.Logger.Error("Ошибка при обновлении статуса попытки авторизации",
				zap.String("auth_attempt_id", attempt.ID), zap.Error(err))
			updateErrors = append(updateErrors, err)
		}
	}

	// Если были ошибки обновления, логируем их
	if len(updateErrors) > 0 {
		logger.Logger.Error("Ошибки при обновлении статусов попыток авторизации", zap.Int("количество", len(updateErrors)))
		return fmt.Errorf("ошибки обновления статусов: %d", len(updateErrors))
	}

	logger.Logger.Info("Успешно обновлены все истекшие попытки авторизации")
	return nil
}
