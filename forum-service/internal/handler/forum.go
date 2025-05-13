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
