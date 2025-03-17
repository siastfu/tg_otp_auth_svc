package grpc

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"tg_otp_auth_svc/pkg/logger"

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
	id, err := uuid.Parse(req.Uuid)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format")
	}

	authAttempt, err := h.AuthUC.CheckAuthLink(id)
	if err != nil {
		if errors.Is(err, domain.ErrAuthAttemptNotFound) {
			return nil, status.Errorf(codes.NotFound, "auth attempt not found")
		}
		if errors.Is(err, domain.ErrAuthAttemptNotPending) {
			return nil, status.Errorf(codes.FailedPrecondition, "auth attempt is not pending")
		}
		if errors.Is(err, domain.ErrAuthLinkExpired) {
			return nil, status.Errorf(codes.DeadlineExceeded, "auth link expired")
		}

		logger.Logger.Error("Ошибка в CheckAuthLink", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	return &pb.CheckAuthResponse{
		Success: true,
		Message: "Authorization successful",
	}, nil
}
