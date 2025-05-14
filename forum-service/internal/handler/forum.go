package handler

import (
	"encoding/json"
	"net/http"

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

// @Summary Создать тему
// @Description Создание новой темы форума (только для авторизованных пользователей)
// @Tags topics
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param input body models.TopicRequest true "Данные темы"
// @Success 201 {object} models.Topic
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /topics [post]
func (h *ForumHandler) CreateTopic(w http.ResponseWriter, r *http.Request) {
	var req models.TopicRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Log.Error("Failed to decode request", zap.Error(err))
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	userID, err := utils.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	topic, err := h.service.CreateTopic(userID, req.Title, req.Content)
	if err != nil {
		logger.Log.Error("Failed to create topic", zap.Error(err))
		http.Error(w, "Failed to create topic", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(topic)
}

// @Summary Удалить тему
// @Description Удаление темы (только для администраторов)
// @Tags topics
// @Security ApiKeyAuth
// @Param id path string true "ID темы"
// @Success 204
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /topics/{id} [delete]
func (h *ForumHandler) DeleteTopic(w http.ResponseWriter, r *http.Request) {
	topicID := r.PathValue("id")
	userID, err := utils.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := h.service.DeleteTopic(topicID, userID); err != nil {
		logger.Log.Error("Failed to delete topic", zap.Error(err))
		http.Error(w, "Failed to delete topic", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Summary Получить все темы
// @Description Получение списка всех активных тем
// @Tags topics
// @Produce json
// @Success 200 {array} models.Topic
// @Failure 500 {object} map[string]string
// @Router /topics [get]
func (h *ForumHandler) GetTopics(w http.ResponseWriter, r *http.Request) {
	topics, err := h.service.GetTopics()
	if err != nil {
		logger.Log.Error("Failed to get topics", zap.Error(err))
		http.Error(w, "Failed to get topics", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(topics)
}

// @Summary Получить сообщения
// @Description Получение сообщений по теме или общего чата
// @Tags messages
// @Produce json
// @Param topic_id query string false "ID темы (если нужны сообщения темы)"
// @Success 200 {array} models.Message
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /messages [get]
func (h *ForumHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
	topicID := r.URL.Query().Get("topic_id")
	var messages []*models.Message
	var err error

	if topicID != "" {
		messages, err = h.service.GetTopicMessages(topicID)
	} else {
		messages, err = h.service.GetChatMessages()
	}

	if err != nil {
		logger.Log.Error("Failed to get messages", zap.Error(err))
		http.Error(w, "Failed to get messages", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

// func TestForumHandler_GetTopics(t *testing.T) {
// 	mockSvc := &mocks.ForumService{}
// 	mockSvc.On("GetTopics").Return([]*models.Topic{{ID: "1", Title: "Test"}}, nil)

// 	handler := handler.NewForumHandler(mockSvc)
// 	req := httptest.NewRequest("GET", "/topics", nil)
// 	w := httptest.NewRecorder()

// 	handler.GetTopics(w, req)
// 	assert.Equal(t, http.StatusOK, w.Code)
// 	mockSvc.AssertExpectations(t)
// }
