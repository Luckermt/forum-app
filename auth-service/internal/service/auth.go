package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/luckermt/forum-app/auth-service/internal/repository"
	"github.com/luckermt/forum-app/shared/pkg/logger"
	"github.com/luckermt/forum-app/shared/pkg/models"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(user *models.User) error
	Login(email, password string) (string, error)
	GetUserByID(userID string) (*models.User, error)
	UpdateUser(userID, username, email string) (*models.User, error)
	ValidateToken(token string) (string, error)
	IsUserAdmin(userID string) (bool, error)
	GetUserRole(userID string) (string, bool, error)
}

type authServiceImpl struct {
	repo      repository.Repository
	jwtSecret string
}

func NewAuthService(repo repository.Repository, jwtSecret string) AuthService {
	if logger.Log == nil {
		if err := logger.Init(); err != nil {
			panic("Failed to initialize logger")
		}
	}

	return &authServiceImpl{
		repo:      repo,
		jwtSecret: jwtSecret,
	}
}

func (s *authServiceImpl) Register(user *models.User) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Log.Error("Failed to hash password", zap.Error(err))
		return err
	}

	user.Password = string(hashedPassword)
	user.Role = "user"
	user.CreatedAt = time.Now()
	user.Blocked = false

	return s.repo.CreateUser(user)
}

func (s *authServiceImpl) Login(email, password string) (string, error) {
	user, err := s.repo.GetUserByEmail(email)
	if err != nil {
		logger.Log.Error("Database error",
			zap.String("email", email),
			zap.Error(err))
		return "", ErrInvalidCredentials
	}

	if user.Blocked {
		return "", ErrUserBlocked
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", ErrInvalidCredentials
	}

	token, err := s.generateJWTToken(user)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *authServiceImpl) GetUserByID(userID string) (*models.User, error) {
	return s.repo.GetUserByID(userID)
}

func (s *authServiceImpl) UpdateUser(userID, username, email string) (*models.User, error) {
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	user.Username = username
	user.Email = email

	if err := s.repo.UpdateUser(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *authServiceImpl) generateJWTToken(user *models.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    user.Role,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	return token.SignedString([]byte(s.jwtSecret))
}

func (s *authServiceImpl) ValidateToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if userID, ok := claims["user_id"].(string); ok {
			return userID, nil
		}
	}

	return "", err
}

func (s *authServiceImpl) IsUserAdmin(userID string) (bool, error) {
	role, _, err := s.GetUserRole(userID)
	return role == "admin", err
}

func (s *authServiceImpl) GetUserRole(userID string) (string, bool, error) {
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return "", false, err
	}
	return user.Role, user.Blocked, nil
}

var (
	ErrEmailExists        = errors.New("email already exists")
	ErrUsernameExists     = errors.New("username already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserBlocked        = errors.New("user is blocked")
)
