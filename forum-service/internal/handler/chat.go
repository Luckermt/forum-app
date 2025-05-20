package handler

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/luckermt/forum-app/forum-service/internal/service"
	"github.com/luckermt/forum-app/shared/pkg/logger"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // В production следует ограничить домены
	},
}

type ChatHandler struct {
	service service.ForumService
}

func NewChatHandler(service service.ForumService) *ChatHandler {
	return &ChatHandler{service: service}
}

// @Summary WebSocket соединение для чата
// @Description Установка WebSocket соединения для чата
// @Tags chat
// @Param token query string true "JWT токен"
// @Router /api/ws [get]
func (h *ChatHandler) HandleConnections(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		logger.Log.Warn("WebSocket connection attempt without token")
		return
	}

	userID, err := h.service.ValidateUser(token)
	if err != nil {
		logger.Log.Warn("Invalid token for WebSocket connection",
			zap.Error(err))
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Log.Error("Failed to upgrade to WebSocket",
			zap.Error(err))
		return
	}
	defer conn.Close()

	h.service.RegisterClient(userID, conn)
	defer h.service.UnregisterClient(userID)

	for {
		var msg struct {
			Text string `json:"text"`
		}

		if err := conn.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err) {
				logger.Log.Info("WebSocket closed",
					zap.String("user_id", userID),
					zap.Error(err))
			}
			break
		}

		if err := h.service.HandleChatMessage(userID, msg.Text); err != nil {
			logger.Log.Error("Failed to handle chat message",
				zap.String("user_id", userID),
				zap.Error(err))
		}
	}
}
