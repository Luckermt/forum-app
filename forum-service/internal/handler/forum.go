package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/luckermt/forum-app/forum-service/internal/service"
	"github.com/luckermt/forum-app/shared/pkg/logger"
	"github.com/luckermt/forum-app/shared/pkg/models"
	"github.com/luckermt/forum-app/shared/pkg/utils"
	"go.uber.org/zap"
)

type ForumHandler struct {
	service service.ForumService
}

func NewForumHandler(service service.ForumService) *ForumHandler {
	return &ForumHandler{service: service}
}

// @Summary Получить список тем
// @Description Получение списка тем с пагинацией и поиском
// @Tags topics
// @Produce json
// @Param page query int false "Номер страницы" default(1)
// @Param limit query int false "Количество тем на странице" default(10)
// @Param search query string false "Поисковый запрос"
// @Success 200 {object} models.TopicsResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/topics [get]
func (h *ForumHandler) GetTopics(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 50 {
		limit = 10
	}

	search := r.URL.Query().Get("search")

	topics, total, err := h.service.GetTopics(page, limit, search)
	if err != nil {
		logger.Log.Error("Failed to get topics", zap.Error(err))
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get topics")
		return
	}

	response := models.TopicsResponse{
		Topics: topics,
		Total:  total,
		Page:   page,
		Limit:  limit,
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}

// @Summary Создать тему
// @Description Создание новой темы (только для авторизованных пользователей)
// @Tags topics
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param input body models.TopicRequest true "Данные темы"
// @Success 201 {object} models.Topic
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/topics [post]
func (h *ForumHandler) CreateTopic(w http.ResponseWriter, r *http.Request) {
	var req models.TopicRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Log.Error("Failed to decode request", zap.Error(err))
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	userID, err := utils.GetUserIDFromContext(r.Context())
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	topic, err := h.service.CreateTopic(userID, req.Title, req.Content)
	if err != nil {
		logger.Log.Error("Failed to create topic", zap.Error(err))
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create topic")
		return
	}

	utils.RespondWithJSON(w, http.StatusCreated, topic)
}
func (h *ForumHandler) DeleteTopic(w http.ResponseWriter, r *http.Request) {
	topicID := r.PathValue("id")
	userID, err := utils.GetUserIDFromContext(r.Context())
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	err = h.service.DeleteTopic(topicID, userID)
	if err != nil {
		logger.Log.Error("Failed to delete topic",
			zap.String("topic_id", topicID),
			zap.Error(err))
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to delete topic")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
func (h *ForumHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
	topicID := r.URL.Query().Get("topic_id")
	if topicID == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "topic_id parameter is required")
		return
	}

	messages, err := h.service.GetTopicMessages(topicID)
	if err != nil {
		logger.Log.Error("Failed to get messages",
			zap.String("topic_id", topicID),
			zap.Error(err))
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get messages")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, messages)
}

func (h *ForumHandler) CreateMessage(w http.ResponseWriter, r *http.Request) {
	var req models.MessageRequest
	if err := utils.DecodeRequest(r, &req); err != nil {
		logger.Log.Error("Failed to decode request", zap.Error(err))
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	userID, err := utils.GetUserIDFromContext(r.Context())
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	message := &models.Message{
		TopicID: req.TopicID,
		UserID:  userID,
		Content: req.Content,
		IsChat:  req.IsChat,
	}

	if err := h.service.CreateMessage(message); err != nil {
		logger.Log.Error("Failed to create message",
			zap.String("user_id", userID),
			zap.Error(err))
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create message")
		return
	}

	utils.RespondWithJSON(w, http.StatusCreated, message)
}

