package grpc

import (
	"context"
	"fmt"
	"net"

	"github.com/Luckermt/shared/proto"
	"google.golang.org/grpc"
)

type AuthServer struct {
	server *grpc.Server
	proto.UnimplementedAuthServiceServer
	service AuthService
}

func NewAuthServer(service AuthService) *AuthServer {
	srv := grpc.NewServer()
	authServer := &AuthServer{
		server:  srv,
		service: service,
	}
	proto.RegisterAuthServiceServer(srv, authServer)
	return authServer
}

func (s *AuthServer) Start(port string) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	if err := s.server.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %v", err)
	}
	return nil
}

func (s *AuthServer) Stop() {
	s.server.GracefulStop()
}

// Реализация gRPC методов
func (s *AuthServer) ValidateToken(ctx context.Context, req *proto.TokenRequest) (*proto.TokenResponse, error) {
	userID, err := s.service.ValidateToken(req.Token)
	if err != nil {
		return &proto.TokenResponse{Valid: false}, nil
	}
	return &proto.TokenResponse{
		Valid:  true,
		UserId: userID,
	}, nil
}

func (s *AuthServer) GetUserRole(ctx context.Context, req *proto.UserRequest) (*proto.UserResponse, error) {
	role, blocked, err := s.service.GetUserRole(req.UserId)
	if err != nil {
		return nil, err
	}
	return &proto.UserResponse{
		Role:    role,
		Blocked: blocked,
	}, nil
}
