package handler

import (
	"net/http"

	"github.com/luckermt/forum-app/forum-service/internal/service"
	"github.com/luckermt/forum-app/shared/pkg/logger"
	"github.com/luckermt/forum-app/shared/pkg/utils"
	"go.uber.org/zap"
)

type AdminHandler struct {
	service service.ForumService
}

func NewAdminHandler(service service.ForumService) *AdminHandler {
	return &AdminHandler{service: service}
}

// @Summary Получить список пользователей
// @Description Получение списка пользователей (только для админов)
// @Tags admin
// @Produce json
// @Security ApiKeyAuth
// @Param search query string false "Поисковый запрос"
// @Success 200 {array} models.User
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/users [get]
func (h *AdminHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")

	users, err := h.service.GetUsers(search)
	if err != nil {
		logger.Log.Error("Failed to get users", zap.Error(err))
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get users")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, users)
}

// @Summary Заблокировать пользователя
// @Description Блокировка пользователя (только для админов)
// @Tags admin
// @Security ApiKeyAuth
// @Param id path string true "ID пользователя"
// @Success 204
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/users/{id}/block [post]
func (h *AdminHandler) BlockUser(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	if err := h.service.BlockUser(userID, true); err != nil {
		logger.Log.Error("Failed to block user",
			zap.String("user_id", userID),
			zap.Error(err))
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to block user")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Summary Разблокировать пользователя
// @Description Разблокировка пользователя (только для админов)
// @Tags admin
// @Security ApiKeyAuth
// @Param id path string true "ID пользователя"
// @Success 204
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/users/{id}/unblock [post]
func (h *AdminHandler) UnblockUser(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	if err := h.service.BlockUser(userID, false); err != nil {
		logger.Log.Error("Failed to unblock user",
			zap.String("user_id", userID),
			zap.Error(err))
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to unblock user")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
