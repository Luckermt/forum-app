package service

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/luckermt/forum-app/auth-service/internal/repository"
	"github.com/luckermt/forum-app/shared/pkg/config"
	"github.com/luckermt/forum-app/shared/pkg/logger"
	"github.com/luckermt/forum-app/shared/pkg/models"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// AuthService определяет интерфейс для сервиса аутентификации
type AuthService interface {
	Register(user *models.User) error
	Login(email, password string) (string, error)
	BlockUser(userID string) error
	ValidateToken(token string) (string, error)
	GetUserRole(userID string) (string, bool, error)
}

// authServiceImpl реализует AuthService
type authServiceImpl struct {
	repo repository.Repository
	jwt  config.JWTConfig
}

// NewAuthService создает новый экземпляр AuthService
func NewAuthService(repo repository.Repository, jwt config.JWTConfig) AuthService {
	return &authServiceImpl{
		repo: repo,
		jwt:  jwt,
	}
}

func (s *authServiceImpl) Register(user *models.User) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Log.Error("Failed to hash password", zap.Error(err))
		return err
	}

	user.Password = string(hashedPassword)
	user.Role = "user" // По умолчанию обычный пользователь
	user.CreatedAt = time.Now()
	user.Blocked = false

	return s.repo.CreateUser(user)
}

func (s *authServiceImpl) Login(email, password string) (string, error) {
	user, err := s.repo.GetUserByEmail(email)
	if err != nil {
		return "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	// Генерация JWT токена
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(s.jwt.ExpiresIn).Unix(),
	})

	return token.SignedString([]byte(s.jwt.SecretKey))
}

func (s *authServiceImpl) BlockUser(userID string) error {
	return s.repo.BlockUser(userID)
}

func (s *authServiceImpl) ValidateToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.jwt.SecretKey), nil
	})
	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if userID, ok := claims["user_id"].(string); ok {
			return userID, nil
		}
	}

	return "", errors.New("invalid token")
}

func (s *authServiceImpl) GetUserRole(userID string) (string, bool, error) {
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return "", false, err
	}
	return user.Role, user.Blocked, nil
}

// func TestAuthService_Register(t *testing.T) {
// 	repo := &mocks.Repository{}
// 	jwtCfg := config.JWTConfig{SecretKey: "test", ExpiresIn: time.Hour}

// 	authSvc := service.NewAuthService(repo, jwtCfg)

// 	t.Run("success", func(t *testing.T) {
// 		user := &models.User{
// 			Email:    "test@example.com",
// 			Password: "password",
// 		}

// 		repo.On("CreateUser", user).Return(nil)

// 		err := authSvc.Register(user)
// 		assert.NoError(t, err)
// 		repo.AssertExpectations(t)
// 	})
// }
