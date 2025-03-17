package repository

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"tg_otp_auth_svc/internal/domain"
	"tg_otp_auth_svc/pkg/database"
	"github.com/jackc/pgx/v5"
)

// PostgresUserRepository - реализация UserRepository
type PostgresUserRepository struct{}

func (r *PostgresUserRepository) GetUserByChatID(chatID int64) (*domain.User, error) {
	var user domain.User
	err := database.DB.QueryRow(context.Background(),
		"SELECT id, username, chat_id, full_name, phone_number, locale, created_at, updated_at FROM \"user\".\"profile\" WHERE chat_id=$1", chatID).
		Scan(&user.ID, &user.Username, &user.ChatID, &user.FullName, &user.PhoneNumber, &user.Locale, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *PostgresUserRepository) CreateUser(user *domain.User) error {
	err := database.DB.QueryRow(context.Background(),
		"INSERT INTO user.profile (username, chat_id, full_name, phone_number, locale, created_at, updated_at) "+
			"VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id",
		user.Username, user.ChatID, user.FullName, user.PhoneNumber, user.Locale, user.CreatedAt, user.UpdatedAt).
		Scan(&user.ID) // ✅ Теперь получаем id нового пользователя

	if err != nil {
		logger.Logger.Error("Ошибка при создании пользователя", zap.Error(err))
		return err
	}

	logger.Logger.Info("Пользователь успешно создан", zap.Int("id", user.ID))
	return nil
}

func (r *PostgresUserRepository) UpdateUser(user *domain.User) error {
	_, err := database.DB.Exec(context.Background(),
		"UPDATE \"user\".\"profile\" SET username=$1, full_name=$2, phone_number=$3, updated_at=$4 WHERE chat_id=$5",
		user.Username, user.FullName, user.PhoneNumber, user.UpdatedAt, user.ChatID)
	return err
}

func (r *PostgresUserRepository) UpdateUserLanguage(chatID int64, locale string) error {
	fmt.Println("🔍 DEBUG: Обновление языка, chat_id =", chatID, "locale =", locale)

	_, err := database.DB.Exec(context.Background(),
		"UPDATE \"user\".\"profile\" SET locale=$1, updated_at=NOW() WHERE chat_id=$2",
		locale, chatID)

	if err != nil {
		fmt.Println("❌ DEBUG: Ошибка SQL-запроса в UpdateUserLanguage:", err)
	}
	return err
}

// PostgresAuthAttemptRepository - реализация AuthAttemptRepository
type PostgresAuthAttemptRepository struct{}

var attempt domain.AuthAttempt


func (r *PostgresAuthAttemptRepository) GetAuthAttemptByID(id uuid.UUID) (*domain.AuthAttempt, error) {
	var authAttempt domain.AuthAttempt
	err := database.DB.QueryRow(context.Background(),
		"SELECT id, tg_id, status_id, created_at, expired_at, succeeded_at FROM auth.attempt WHERE id=$1", id).
		Scan(&authAttempt.ID, &authAttempt.TgID, &authAttempt.StatusID, &authAttempt.CreatedAt, &authAttempt.ExpiredAt, &authAttempt.SucceededAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			logger.Logger.Warn("Попытка авторизации не найдена", zap.String("uuid", id.String()))
			return nil, nil // ✅ Теперь возвращаем `nil`, а не ошибку
		}
		return nil, err
	}

	return &authAttempt, nil
}

func (r *PostgresAuthAttemptRepository) CreateAuthAttempt(attempt *domain.AuthAttempt) error {
	_, err := database.DB.Exec(context.Background(),
		"INSERT INTO auth.attempt (id, tg_id, status_id, created_at, expired_at) VALUES ($1, $2, $3, $4, $5)",
		attempt.ID, attempt.TgID, attempt.StatusID, attempt.CreatedAt, attempt.ExpiredAt)
	return err
}

func (r *PostgresAuthAttemptRepository) UpdateAuthAttempt(attempt *domain.AuthAttempt) error {
	_, err := database.DB.Exec(context.Background(),
		"UPDATE auth.attempt SET status_id=$1, succeeded_at=$2 WHERE id=$3",
		attempt.StatusID, attempt.SucceededAt, attempt.ID)
	return err
}

// ✅ Новый метод: Получение истекших попыток авторизации
func (r *PostgresAuthAttemptRepository) GetExpiredAuthAttempts() ([]domain.AuthAttempt, error) {
	rows, err := database.DB.Query(context.Background(),
		`SELECT id, tg_id, status_id, created_at, expired_at, succeeded_at 
		 FROM auth.attempt 
		 WHERE status_id = 1 AND expired_at <= NOW()`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attempts []domain.AuthAttempt
	for rows.Next() {
		var attempt domain.AuthAttempt
		err := rows.Scan(&attempt.ID, &attempt.TgID, &attempt.StatusID, &attempt.CreatedAt, &attempt.ExpiredAt, &attempt.SucceededAt)
		if err != nil {
			return nil, err
		}
		attempts = append(attempts, attempt)
	}

	return attempts, nil
}

func (r *PostgresAuthAttemptRepository) UpdateAuthAttemptStatus(authID uuid.UUID, statusID int) error {
	_, err := database.DB.Exec(context.Background(),
		"UPDATE auth.attempt SET status_id=$1 WHERE id=$2",
		statusID, authID)

	if err != nil {
		logger.Logger.Error("Ошибка при обновлении статуса авторизации", zap.Error(err), zap.String("authID", authID.String()))
		return err
	}

	logger.Logger.Info("Статус авторизации обновлён", zap.String("authID", authID.String()), zap.Int("statusID", statusID))
	return nil
}

// PostgresAuthAttemptStatusRepository - реализация AuthAttemptStatusRepository
type PostgresAuthAttemptStatusRepository struct{}

func (r *PostgresAuthAttemptStatusRepository) GetStatusByID(id int) (*domain.AuthAttemptStatus, error) {
	var status domain.AuthAttemptStatus
	err := database.DB.QueryRow(context.Background(),
		"SELECT id, tech_name, display_name::text, display_image::text FROM list.auth_attempt_status WHERE id=$1", id).
		Scan(&status.ID, &status.TechName, &status.DisplayName, &status.DisplayImage)
	if err != nil {
		return nil, err
	}
	return &status, nil
}

func (r *PostgresAuthAttemptRepository) GetPendingAuthAttemptByTgID(tgID int64) (*domain.AuthAttempt, error) {
	var attempt domain.AuthAttempt
	err := database.DB.QueryRow(context.Background(),
		"SELECT id, tg_id, status_id, created_at, expired_at, succeeded_at FROM \"auth\".\"attempt\" WHERE tg_id=$1 AND status_id=1",
		tgID).Scan(&attempt.ID, &attempt.TgID, &attempt.StatusID, &attempt.CreatedAt, &attempt.ExpiredAt, &attempt.SucceededAt)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, nil
		}
		return nil, err
	}

	return &attempt, nil
}

// GetUserByTgID - получает пользователя по Telegram ID
func (r *PostgresAuthAttemptRepository) GetUserByTgID(tgID int64) (*domain.User, error) {
	var user domain.User
	err := database.DB.QueryRow(context.Background(),
		"SELECT id, username, tg_id, full_name, locale FROM user.profile WHERE tg_id=$1", tgID).
		Scan(&user.ID, &user.Username, &user.ChatID, &user.FullName, &user.Locale)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// CreateUser - создает нового пользователя
func (r *PostgresAuthAttemptRepository) CreateUser(user *domain.User) error {
	_, err := database.DB.Exec(context.Background(),
		"INSERT INTO user.profile (username, tg_id, full_name, locale) VALUES ($1, $2, $3, $4)",
		user.Username, user.ChatID, user.FullName, user.Locale)
	return err
}

// UpdateUserLanguage - обновляет язык пользователя
func (r *PostgresAuthAttemptRepository) UpdateUserLanguage(tgID int64, lang string) error {
	_, err := database.DB.Exec(context.Background(),
		"UPDATE user.profile SET locale=$1 WHERE tg_id=$2", lang, tgID)
	return err
}
import "github.com/jackc/pgx/v5"

func (r *PostgresAuthAttemptRepository) DeleteExpiredAuthAttempts() error {
	query := "DELETE FROM auth.attempt WHERE status_id = 2 AND expired_at < NOW()"

	result, err := database.DB.Exec(context.Background(), query)
	if err != nil {
		logger.Logger.Error("Ошибка при удалении истёкших попыток авторизации", zap.Error(err))
		return err
	}

	rowsDeleted := result.RowsAffected()
	logger.Logger.Info("Удалено истёкших попыток авторизации", zap.Int64("rows", rowsDeleted))

	return nil
}
