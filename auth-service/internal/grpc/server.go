package grpc

import (
	"context"

	"github.com/luckermt/shared/proto"
)

type AuthServer struct {
	service AuthService
	proto.UnimplementedAuthServiceServer
}

func NewAuthServer(service AuthService) *AuthServer {
	return &AuthServer{service: service}
}

func (s *AuthServer) ValidateToken(ctx context.Context, req *proto.TokenRequest) (*proto.TokenResponse, error) {
	// Реализация валидации токена
}

func (s *AuthServer) GetUserRole(ctx context.Context, req *proto.UserRequest) (*proto.UserResponse, error) {
	// Получение роли пользователя
}
