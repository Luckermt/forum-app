package handler

import (
	"encoding/json"
	"net/http"

	"github.com/luckermt/forum-app/auth-service/internal/service"
	"github.com/luckermt/forum-app/shared/pkg/logger"
	"github.com/luckermt/forum-app/shared/pkg/models"
	"github.com/luckermt/forum-app/shared/pkg/utils"
	"go.uber.org/zap"
)

type UserHandler struct {
	service service.AuthService
}

func NewUserHandler(service service.AuthService) *UserHandler {
	return &UserHandler{service: service}
}

// @Summary Получить информацию о пользователе
// @Description Получение информации о конкретном пользователе
// @Tags users
// @Produce json
// @Param id path string true "ID пользователя"
// @Success 200 {object} models.UserResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/users/{id} [get]
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")

	user, err := h.service.GetUserByID(userID)
	if err != nil {
		logger.Log.Error("Failed to get user",
			zap.String("user_id", userID),
			zap.Error(err))
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get user")
		return
	}

	response := models.UserResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
		Blocked:  user.Blocked,
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}

// @Summary Обновить информацию о пользователе
// @Description Обновление информации о пользователе (только для самого пользователя или админа)
// @Tags users
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "ID пользователя"
// @Param input body models.UpdateUserRequest true "Данные для обновления"
// @Success 200 {object} models.UserResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/users/{id} [put]
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	currentUserID, err := utils.GetUserIDFromContext(r.Context())
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Проверка прав доступа
	isAdmin, err := h.service.IsUserAdmin(currentUserID)
	if err != nil {
		logger.Log.Error("Failed to check admin status",
			zap.String("user_id", currentUserID),
			zap.Error(err))
		utils.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	if currentUserID != userID && !isAdmin {
		utils.RespondWithError(w, http.StatusForbidden, "Forbidden")
		return
	}

	var req models.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Log.Error("Failed to decode request",
			zap.Error(err))
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	user, err := h.service.UpdateUser(userID, req.Username, req.Email)
	if err != nil {
		logger.Log.Error("Failed to update user",
			zap.String("user_id", userID),
			zap.Error(err))
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update user")
		return
	}

	response := models.UserResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
		Blocked:  user.Blocked,
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}
