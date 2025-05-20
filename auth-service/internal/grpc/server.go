package grpc

import (
	"context"
	"fmt"
	"net"

	"github.com/luckermt/forum-app/auth-service/internal/service"
	"github.com/luckermt/forum-app/shared/pkg/logger"
	"github.com/luckermt/forum-app/shared/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// AuthServer реализует gRPC сервер для аутентификации
type AuthServer struct {
	grpcServer  *grpc.Server
	authService service.AuthService
	proto.UnimplementedAuthServiceServer
}

// NewAuthServer создает новый экземпляр AuthServer
func NewAuthServer(authService service.AuthService) *AuthServer {
	srv := grpc.NewServer()
	server := &AuthServer{
		grpcServer:  srv,
		authService: authService,
	}
	proto.RegisterAuthServiceServer(srv, server)
	return server
}

// Start запускает gRPC сервер
func (s *AuthServer) Start(port string) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	logger.Log.Info("Starting gRPC server",
		zap.String("port", port),
		zap.String("service", "auth"))

	return s.grpcServer.Serve(lis)
}

// Stop останавливает gRPC сервер
func (s *AuthServer) Stop() {
	if s.grpcServer != nil {
		logger.Log.Info("Gracefully stopping gRPC server")
		s.grpcServer.GracefulStop()
	}
}

// ValidateToken реализует gRPC метод проверки токена
func (s *AuthServer) ValidateToken(ctx context.Context, req *proto.TokenRequest) (*proto.TokenResponse, error) {
	userID, err := s.authService.ValidateToken(req.Token)
	if err != nil {
		return &proto.TokenResponse{Valid: false}, nil
	}
	return &proto.TokenResponse{
		Valid:  true,
		UserId: userID,
	}, nil
}

// GetUserRole реализует gRPC метод получения роли пользователя
func (s *AuthServer) GetUserRole(ctx context.Context, req *proto.UserRequest) (*proto.UserResponse, error) {
	role, blocked, err := s.authService.GetUserRole(req.UserId)
	if err != nil {
		return nil, err
	}
	return &proto.UserResponse{
		Role:    role,
		Blocked: blocked,
	}, nil
}

// GetUserInfo реализует gRPC метод получения информации о пользователе
func (s *AuthServer) GetUserInfo(ctx context.Context, req *proto.UserRequest) (*proto.UserInfoResponse, error) {
	user, err := s.authService.GetUserByID(req.UserId)
	if err != nil {
		return nil, err
	}
	return &proto.UserInfoResponse{
		Id:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
	}, nil
}
