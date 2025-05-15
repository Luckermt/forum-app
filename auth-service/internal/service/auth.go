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

// AuthService определяет контракт сервиса аутентификации
type AuthService interface {
	Register(user *models.User) error
	Login(email, password string) (string, error)
	ValidateToken(token string) (string, error)
	GetUserRole(userID string) (string, bool, error)
}

type authServiceImpl struct {
	repo      repository.Repository
	jwtSecret string
}

func NewAuthService(repo repository.Repository, jwtSecret string) AuthService {
	// Инициализация логгера при создании сервиса
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
	// Проверка инициализации логгера
	if logger.Log == nil {
		return errors.New("logger not initialized")
	}

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

	logger.Log.Debug("Login attempt",
		zap.String("email", email),
		zap.Time("timestamp", time.Now()),
	)

	user, err := s.repo.GetUserByEmail(email)
	if err != nil {
		logger.Log.Error("Database error",
			zap.Error(err),
			zap.String("email", email),
		)
		return "", fmt.Errorf("internal server error")
	}

	if user == nil {
		logger.Log.Warn("User not found", zap.String("email", email))
		return "", ErrInvalidCredentials
	}

	logger.Log.Debug("User data from DB",
		zap.String("db_email", user.Email),
		zap.String("db_pwd_prefix", user.Password[:10]),
		zap.Bool("blocked", user.Blocked),
	)

	if user.Blocked {
		logger.Log.Warn("Blocked user attempt", zap.String("email", email))
		return "", ErrUserBlocked
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		logger.Log.Warn("Password mismatch",
			zap.String("input_pwd", maskPassword(password)),
			zap.String("db_pwd_prefix", user.Password[:10]),
			zap.Error(err),
		)
		return "", ErrInvalidCredentials
	}

	logger.Log.Info("Successful login", zap.String("email", email))
	return s.generateJWTToken(user)
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

func (s *authServiceImpl) GetUserRole(userID string) (string, bool, error) {
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return "", false, err
	}
	return user.Role, user.Blocked, nil
}
func maskPassword(pwd string) string {
	if len(pwd) == 0 {
		return ""
	}
	if len(pwd) == 1 {
		return "*"
	}
	return string(pwd[0]) + "***" + string(pwd[len(pwd)-1])
}

var (
	ErrEmailExists        = errors.New("email already exists")
	ErrUsernameExists     = errors.New("username already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserBlocked        = errors.New("user is blocked")
)
