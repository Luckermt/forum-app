package handler

import (
	"encoding/json"
	"net/http"

	"github.com/luckermt/shared/pkg/logger"
	"github.com/luckermt/shared/pkg/models"
	"go.uber.org/zap"
)

type AuthHandler struct {
	service AuthService
}

func NewAuthHandler(service AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

// Register godoc
// @Summary Регистрация нового пользователя
// @Description Создает учетную запись пользователя
// @Tags auth
// @Accept  json
// @Produce  json
// @Param input body models.User true "Данные пользователя"
// @Success 201 {object} models.User
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		logger.Log.Error("Failed to decode request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.Register(&user); err != nil {
		logger.Log.Error("Failed to register user", zap.Error(err))
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	user.Password = "" // Не возвращаем пароль в ответе
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

// Login godoc
// @Summary Авторизация пользователя
// @Description Вход в систему с email и паролем
// @Tags auth
// @Accept  json
// @Produce  json
// @Param input body models.AuthRequest true "Учетные данные"
// @Success 200 {object} models.AuthResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var creds models.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		logger.Log.Error("Failed to decode request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	token, err := h.service.Login(creds.Email, creds.Password)
	if err != nil {
		logger.Log.Error("Login failed", zap.Error(err))
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	response := models.AuthResponse{Token: token}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// BlockUser godoc
// @Summary Блокировка пользователя (Admin)
// @Description Блокирует пользователя по ID
// @Tags admin
// @Security ApiKeyAuth
// @Param id path string true "User ID"
// @Success 204
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/block/{id} [post]
func (h *AuthHandler) BlockUser(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	if err := h.service.BlockUser(userID); err != nil {
		logger.Log.Error("Failed to block user", zap.Error(err))
		http.Error(w, "Failed to block user", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
