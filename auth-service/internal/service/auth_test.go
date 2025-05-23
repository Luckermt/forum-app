package service_test

import (
	"testing"

	"github.com/luckermt/forum-app/auth-service/internal/repository/mocks"
	"github.com/luckermt/forum-app/auth-service/internal/service"
	"github.com/luckermt/forum-app/shared/pkg/logger"
	"github.com/luckermt/forum-app/shared/pkg/models"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthService_Register(t *testing.T) {

	if err := logger.Init(); err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Log.Sync()

	repo := new(mocks.Repository)

	jwtSecret := "test-secret-key"

	authSvc := service.NewAuthService(repo, jwtSecret)

	user := &models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "SecurePass123!",
	}

	repo.On("CreateUser", user).Return(nil)

	err := authSvc.Register(user)

	// Проверки
	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestAuthService_Login(t *testing.T) {

	if err := logger.Init(); err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Log.Sync()

	repo := new(mocks.Repository)

	jwtSecret := "test-secret-key"
	email := "test@example.com"
	password := "correct_password"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	authSvc := service.NewAuthService(repo, jwtSecret)

	repo.On("GetUserByEmail", email).Return(&models.User{
		Email:    email,
		Password: string(hashedPassword),
		Blocked:  false,
	}, nil)

	token, err := authSvc.Login(email, password)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	repo.AssertExpectations(t)
}
