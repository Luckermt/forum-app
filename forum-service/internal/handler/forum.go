package handler

import (
	"net/http"

	"github.com/luckermt/shared/pkg/logger"
	"github.com/luckermt/shared/pkg/models"
	"go.uber.org/zap"
)

type ForumHandler struct {
	service *ForumService
}

func NewForumHandler(service *ForumService) *ForumHandler {
	return &ForumHandler{service: service}
}

func (h *ForumHandler) HandleTopics(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getTopics(w, r)
	case http.MethodPost:
		h.createTopic(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *ForumHandler) getTopics(w http.ResponseWriter, r *http.Request) {
	topics, err := h.service.GetTopics()
	if err != nil {
		logger.Log.Error("Failed to get topics", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, topics)
}

func (h *ForumHandler) createTopic(w http.ResponseWriter, r *http.Request) {
	var req models.TopicRequest
	if err := decodeRequest(r, &req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	userID, err := getUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	topic, err := h.service.CreateTopic(userID, req.Title, req.Content)
	if err != nil {
		logger.Log.Error("Failed to create topic", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusCreated, topic)
}
