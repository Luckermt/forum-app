package service

import (
	"testing"
	"time"

	"github.com/luckermt/forum-app/auth-service/internal/repository/mocks"
	"github.com/luckermt/forum-app/shared/pkg/config"
	"github.com/luckermt/forum-app/shared/pkg/models"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthService_Register(t *testing.T) {
	repo := new(mocks.Repository)
	jwtCfg := config.JWTConfig{
		SecretKey: "test-secret",
		ExpiresIn: time.Hour,
	}

	authSvc := NewAuthService(repo, jwtCfg)

	t.Run("successful registration", func(t *testing.T) {
		user := &models.User{
			Email:    "test@example.com",
			Password: "password123",
		}

		repo.On("CreateUser", user).Return(nil)

		err := authSvc.Register(user)
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("failed registration (user already exists)", func(t *testing.T) {
		user := &models.User{
			Email:    "existing@example.com",
			Password: "password123",
		}

		repo.On("CreateUser", user).Return(assert.AnError)

		err := authSvc.Register(user)
		assert.Error(t, err)
		repo.AssertExpectations(t)
	})
}

func TestAuthService_Login(t *testing.T) {
	repo := new(mocks.Repository)
	jwtCfg := config.JWTConfig{
		SecretKey: "test-secret",
		ExpiresIn: time.Hour,
	}

	authSvc := NewAuthService(repo, jwtCfg)

	t.Run("successful login", func(t *testing.T) {
		email := "test@example.com"
		password := "password123"
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

		repo.On("GetUserByEmail", email).Return(&models.User{
			Email:    email,
			Password: string(hashedPassword),
		}, nil)

		token, err := authSvc.Login(email, password)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
		repo.AssertExpectations(t)
	})

	t.Run("failed login (invalid password)", func(t *testing.T) {
		email := "test@example.com"
		wrongPassword := "wrong-password"

		repo.On("GetUserByEmail", email).Return(&models.User{
			Email:    email,
			Password: "hashed-password",
		}, nil)

		_, err := authSvc.Login(email, wrongPassword)
		assert.Error(t, err)
		repo.AssertExpectations(t)
	})
}
