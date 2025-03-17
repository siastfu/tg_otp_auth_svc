package grpc

import (
	"context"
	"errors"
	"time"

	"tg_otp_auth_svc/internal/domain"
	pb "tg_otp_auth_svc/proto" // gRPC-протофайл
)

// AuthHandler - структура хендлера gRPC
type AuthHandler struct {
	pb.UnimplementedAuthServiceServer // Добавляем это поле
	AuthUC                            domain.AuthUseCase
}

// Убеждаемся, что AuthHandler реализует интерфейс pb.AuthServiceServer
var _ pb.AuthServiceServer = (*AuthHandler)(nil)

// NewAuthHandler - создаёт новый экземпляр AuthHandler
func NewAuthHandler(authUC domain.AuthUseCase) *AuthHandler {
	return &AuthHandler{AuthUC: authUC}
}

// CheckAuthLink - обработка gRPC-запроса на проверку ссылки
func (h *AuthHandler) CheckAuthLink(ctx context.Context, req *pb.CheckAuthRequest) (*pb.CheckAuthResponse, error) {
	authAttempt, err := h.AuthUC.CheckAuthLink(req.Uuid)
	if err != nil {
		if errors.Is(err, domain.ErrAuthAttemptNotFound) {
			return nil, errors.New("auth attempt not found")
		}
		if errors.Is(err, domain.ErrAuthAttemptNotPending) {
			return nil, errors.New("auth attempt is not pending")
		}
		if errors.Is(err, domain.ErrAuthLinkExpired) {
			return nil, errors.New("auth link expired")
		}
		return nil, err
	}

	// Обновляем статус авторизации в БД
	authAttempt.StatusID = 3 // SUCCESS
	now := time.Now()
	authAttempt.SucceededAt = &now

	err = h.AuthUC.UpdateAuthAttempt(authAttempt)
	if err != nil {
		return nil, err
	}

	// Возвращаем успешный ответ
	return &pb.CheckAuthResponse{
		TgId:    authAttempt.TgID,
		Success: true,
	}, nil
}
