package service

import (
	"context"

	"github.com/luckermt/forum-app/shared/pkg/models"
	"github.com/luckermt/forum-app/shared/proto"
)

// type AuthClient interface {
// 	ValidateToken(token string) (string, error)
// 	IsUserAdmin(userID string) (bool, error)
// }

type GRPCAuthClient struct {
	client proto.AuthServiceClient
}

func NewGRPCAuthClient(client proto.AuthServiceClient) AuthClient {
	return &GRPCAuthClient{client: client}
}

func (c *GRPCAuthClient) ValidateToken(token string) (string, error) {
	resp, err := c.client.ValidateToken(context.Background(), &proto.TokenRequest{Token: token})
	if err != nil {
		return "", err
	}
	if !resp.Valid {
		return "", ErrInvalidToken
	}
	return resp.UserId, nil
}

func (c *GRPCAuthClient) IsUserAdmin(userID string) (bool, error) {
	resp, err := c.client.GetUserRole(context.Background(), &proto.UserRequest{UserId: userID})
	if err != nil {
		return false, err
	}
	return resp.Role == "admin", nil
}
func (c *GRPCAuthClient) GetUserInfo(userID string) (*models.User, error) {
	resp, err := c.client.GetUserInfo(context.Background(), &proto.UserRequest{UserId: userID})
	if err != nil {
		return nil, err
	}

	return &models.User{
		ID:       resp.Id,
		Username: resp.Username,
		Email:    resp.Email,
		Role:     resp.Role,
	}, nil
}
