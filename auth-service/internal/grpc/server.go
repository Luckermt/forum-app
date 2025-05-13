package grpc

import (
	"context"
	"fmt"
	"net"

	"github.com/luckermt/forum-app/auth-service/internal/service"
	"github.com/luckermt/forum-app/shared/proto"
	"google.golang.org/grpc"
)

// AuthServer реализует gRPC сервер для аутентификации
type AuthServer struct {
	proto.UnimplementedAuthServiceServer
	authService service.AuthService
}

// NewAuthServer создает новый экземпляр AuthServer
func NewAuthServer(authService service.AuthService) *AuthServer {
	return &AuthServer{
		authService: authService,
	}
}

// Start запускает gRPC сервер
func (s *AuthServer) Start(port string) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	server := grpc.NewServer()
	proto.RegisterAuthServiceServer(server, s)

	if err := server.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %v", err)
	}
	return nil
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
