package grpc

import (
	"context"
	"errors"

	"github.com/luckermt/forum-app/shared/proto"
	"google.golang.org/grpc"
)

type AuthClient struct {
	conn   *grpc.ClientConn
	client proto.AuthServiceClient
}

func NewAuthClient(addr string) (*AuthClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	return &AuthClient{
		conn:   conn,
		client: proto.NewAuthServiceClient(conn),
	}, nil
}

func (c *AuthClient) ValidateToken(token string) (string, error) {
	resp, err := c.client.ValidateToken(context.Background(), &proto.TokenRequest{Token: token})
	if err != nil {
		return "", err
	}

	if !resp.Valid {
		return "", errors.New("invalid token")
	}

	return resp.UserId, nil
}
func (c *AuthClient) IsUserAdmin(userID string) (bool, error) {
	resp, err := c.client.GetUserRole(context.Background(), &proto.UserRequest{
		UserId: userID,
	})
	if err != nil {
		return false, err
	}
	return resp.Role == "admin", nil
}
