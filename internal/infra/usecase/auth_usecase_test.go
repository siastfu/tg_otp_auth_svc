package usecase

import (
	"errors"
	"github.com/google/uuid"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"tg_otp_auth_svc/internal/domain"
)

// Mock-репозиторий для тестирования
type MockAuthRepo struct {
	mock.Mock
}

func (m *MockAuthRepo) GetAuthAttemptByID(id uuid.UUID) (*domain.AuthAttempt, error) {
	args := m.Called(id)

	// Проверяем, что значение не nil
	if attempt, ok := args.Get(0).(*domain.AuthAttempt); ok {
		return attempt, args.Error(1)
	}

	return nil, args.Error(1) // ✅ Если nil, возвращаем nil без panic
}

func (m *MockAuthRepo) GetPendingAuthAttemptByTgID(tgID int64) (*domain.AuthAttempt, error) {
	args := m.Called(tgID)
	return args.Get(0).(*domain.AuthAttempt), args.Error(1)
}

func (m *MockAuthRepo) CreateAuthAttempt(attempt *domain.AuthAttempt) error {
	args := m.Called(attempt)
	return args.Error(0)
}

func (m *MockAuthRepo) UpdateAuthAttempt(attempt *domain.AuthAttempt) error {
	args := m.Called(attempt)
	return args.Error(0)
}

func (m *MockAuthRepo) GetExpiredAuthAttempts() ([]domain.AuthAttempt, error) {
	args := m.Called()
	return args.Get(0).([]domain.AuthAttempt), args.Error(1)
}

func (m *MockAuthRepo) UpdateAuthAttemptStatus(authID string, statusID int) error {
	args := m.Called(authID, statusID)
	return args.Error(0)
}

// Тестируем MarkExpiredAuthAttempts()
func TestMarkExpiredAuthAttempts(t *testing.T) {
	mockRepo := new(MockAuthRepo)
	authUC := NewAuthUseCase(mockRepo)

	// Тестовый набор данных
	expiredAttempts := []domain.AuthAttempt{
		{ID: "1", TgID: 123, StatusID: 1, CreatedAt: time.Now(), ExpiredAt: time.Now().Add(-1 * time.Minute)},
		{ID: "2", TgID: 456, StatusID: 1, CreatedAt: time.Now(), ExpiredAt: time.Now().Add(-2 * time.Minute)},
	}

	// Настраиваем mock-методы
	mockRepo.On("GetExpiredAuthAttempts").Return(expiredAttempts, nil)
	mockRepo.On("UpdateAuthAttemptStatus", "1", 2).Return(nil)
	mockRepo.On("UpdateAuthAttemptStatus", "2", 2).Return(nil)

	// Вызываем метод
	err := authUC.MarkExpiredAuthAttempts()

	// Проверяем результат
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// Тест: если нет истекших попыток
func TestMarkExpiredAuthAttempts_NoAttempts(t *testing.T) {
	mockRepo := new(MockAuthRepo)
	authUC := NewAuthUseCase(mockRepo)

	// Настраиваем mock-методы
	mockRepo.On("GetExpiredAuthAttempts").Return([]domain.AuthAttempt{}, nil)

	// Вызываем метод
	err := authUC.MarkExpiredAuthAttempts()

	// Проверяем, что ошибки нет
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// Тест: если БД вернула ошибку
func TestMarkExpiredAuthAttempts_DBError(t *testing.T) {
	mockRepo := new(MockAuthRepo)
	authUC := NewAuthUseCase(mockRepo)

	// Настраиваем ошибку от базы
	mockRepo.On("GetExpiredAuthAttempts").Return(nil, errors.New("database error"))

	// Вызываем метод
	err := authUC.MarkExpiredAuthAttempts()

	// Проверяем, что вернулась ошибка
	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}
