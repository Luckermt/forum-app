package handler

import (
	"forum-service/internal/service"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/luckermt/shared/pkg/logger"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type ChatHandler struct {
	service *service.ForumService
}

func NewChatHandler(service *service.ForumService) *ChatHandler {
	return &ChatHandler{service: service}
}

func (h *ChatHandler) HandleConnections(w http.ResponseWriter, r *http.Request) {
	// Проверка аутентификации
	token := r.URL.Query().Get("token")
	userID, err := h.service.ValidateUser(token)
	if err != nil {
		logger.Log.Error("Unauthorized websocket connection", zap.Error(err))
		return
	}

	// Обновление соединения до WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Log.Error("Failed to upgrade to websocket", zap.Error(err))
		return
	}
	defer conn.Close()

	// Регистрация клиента
	h.service.RegisterClient(userID, conn)

	for {
		var msg struct {
			Text string `json:"text"`
		}
		err := conn.ReadJSON(&msg)
		if err != nil {
			logger.Log.Error("Error reading message", zap.Error(err))
			h.service.UnregisterClient(userID)
			break
		}

		// Сохранение и рассылка сообщения
		if err := h.service.HandleChatMessage(userID, msg.Text); err != nil {
			logger.Log.Error("Error handling message", zap.Error(err))
		}
	}
}
