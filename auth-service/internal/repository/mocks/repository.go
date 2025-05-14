package mocks

import (
	"github.com/luckermt/forum-app/shared/pkg/models"
	"github.com/stretchr/testify/mock"
)

type Repository struct {
	mock.Mock
}

func (m *Repository) CreateUser(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *Repository) GetUserByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *Repository) GetUserByID(userID string) (*models.User, error) {
	args := m.Called(userID)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *Repository) BlockUser(userID string) error {
	args := m.Called(userID)
	return args.Error(0)
}
