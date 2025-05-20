package repository

import (
	"github.com/luckermt/forum-app/shared/pkg/models"
)

type Repository interface {
	CreateUser(user *models.User) error
	GetUserByEmail(email string) (*models.User, error)
	GetUserByID(userID string) (*models.User, error)
	UpdateUser(user *models.User) error
	BlockUser(userID string, blocked bool) error
}
