package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/luckermt/forum-app/auth-service/internal/service"
	"github.com/luckermt/forum-app/shared/pkg/logger"
	"github.com/luckermt/forum-app/shared/pkg/models"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// RegisterRequest модель запроса регистрации
// swagger:model
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50" example:"john_doe"`
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required,min=8,max=72" example:"SecurePass123!"`
}

// LoginRequest модель запроса входа
// swagger:model
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required" example:"SecurePass123!"`
}

// LoginResponse модель ответа на вход
// swagger:model
type LoginResponse struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// UserResponse модель ответа с данными пользователя
// swagger:model
type UserResponse struct {
	ID        string    `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Username  string    `json:"username" example:"john_doe"`
	Email     string    `json:"email" example:"user@example.com"`
	Role      string    `json:"role" example:"user"`
	CreatedAt time.Time `json:"created_at" example:"2023-01-01T15:04:05Z"`
}

// ErrorResponse модель ошибки
// swagger:model
type ErrorResponse struct {
	Message string `json:"message" example:"error message"`
}

// Register обработчик регистрации пользователя
// @Summary Регистрация нового пользователя
// @Description Создает учетную запись пользователя
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Данные для регистрации"
// @Success 201 {object} UserResponse
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Log.Error("Failed to decode request", zap.Error(err))
		writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := validatePassword(req.Password); err != nil {
		logger.Log.Warn("Password validation failed", zap.Error(err))
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Log.Error("Failed to hash password", zap.Error(err))
		writeError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	user := models.User{
		ID:        uuid.New().String(),
		Username:  req.Username,
		Email:     req.Email,
		Password:  string(hashedPassword),
		Role:      "user",
		CreatedAt: time.Now(),
		Blocked:   false,
	}

	if err := h.authService.Register(&user); err != nil {
		handleServiceError(w, err)
		return
	}

	response := UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// Login обработчик входа пользователя
// @Summary Вход пользователя
// @Description Аутентификация по email и паролю
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Данные для входа"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Log.Error("Failed to decode request",
			zap.Error(err),
			zap.String("remote_addr", r.RemoteAddr),
		)
		writeError(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	token, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		logger.Log.Warn("Login attempt failed",
			zap.String("email", req.Email),
			zap.Error(err),
		)

		switch {
		case errors.Is(err, service.ErrInvalidCredentials):
			writeError(w, "Invalid email or password", http.StatusUnauthorized)
		case errors.Is(err, service.ErrUserBlocked):
			writeError(w, "Account is blocked", http.StatusForbidden)
		default:
			writeError(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	logger.Log.Info("User logged in",
		zap.String("email", req.Email),
	)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(LoginResponse{
		Token: token,
	})
}
func validatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}
	return nil
}

func writeError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{Message: message})
}

func handleServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrEmailExists):
		writeError(w, "Email already registered", http.StatusConflict)
	case errors.Is(err, service.ErrUsernameExists):
		writeError(w, "Username already taken", http.StatusConflict)
	default:
		logger.Log.Error("Registration failed", zap.Error(err))
		writeError(w, "Internal server error", http.StatusInternalServerError)
	}
}
func (h *AuthHandler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			writeError(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		userID, err := h.authService.ValidateToken(tokenString)
		if err != nil {
			writeError(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "userID", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
