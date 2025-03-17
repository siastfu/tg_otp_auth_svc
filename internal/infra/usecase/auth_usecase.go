package usecase

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"tg_otp_auth_svc/internal/domain"
)

// AuthUseCaseImpl - реализация AuthUseCase
type AuthUseCaseImpl struct {
	AuthRepo domain.AuthAttemptRepository
}

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
		return nil, errors.New("auth attempt not found")
	}

	if authAttempt.StatusID != 1 {
		return nil, errors.New("auth attempt is not pending")
	}

	if time.Now().After(authAttempt.ExpiredAt) {
		return nil, errors.New("auth link expired")
	}

	return authAttempt, nil
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

// UpdateAuthAttempt обновляет статус авторизационной попытки
func (u *AuthUseCaseImpl) UpdateAuthAttempt(attempt *domain.AuthAttempt) error {
	return u.AuthRepo.UpdateAuthAttempt(attempt)
}
