package service

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/luckermt/shared/pkg/config"
	"github.com/luckermt/shared/pkg/logger"
	"github.com/luckermt/shared/pkg/models"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo Repository
	jwt  config.JWTConfig
}

func NewAuthService(repo Repository, jwt config.JWTConfig) *AuthService {
	return &AuthService{repo: repo, jwt: jwt}
}

func (s *AuthService) Register(user *models.User) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Log.Error("Failed to hash password", zap.Error(err))
		return err
	}

	user.Password = string(hashedPassword)
	user.Role = "user" // По умолчанию обычный пользователь

	return s.repo.CreateUser(user)
}

func (s *AuthService) Login(email, password string) (string, error) {
	user, err := s.repo.GetUserByEmail(email)
	if err != nil {
		return "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", err
	}

	// Генерация JWT токена
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(s.jwt.ExpiresIn).Unix(),
	})

	return token.SignedString([]byte(s.jwt.SecretKey))
}

func (s *AuthService) BlockUser(userID string) error {
	return s.repo.BlockUser(userID)
}
