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
		return true // В production замените на проверку origin
	},
}

type ChatHandler struct {
	service service.ForumService
}

func NewChatHandler(service service.ForumService) *ChatHandler {
	return &ChatHandler{service: service}
}

// @Summary WebSocket чат
// @Description Подключение к чату через WebSocket
// @Tags chat
// @Param token query string true "JWT токен"
// @Router /ws [get]
func (h *ChatHandler) HandleConnections(w http.ResponseWriter, r *http.Request) {
	// Проверка аутентификации
	token := r.URL.Query().Get("token")
	userID, err := h.service.ValidateUser(token)
	if err != nil {
		logger.Log.Error("Unauthorized websocket connection",
			zap.Error(err))
		return
	}

	// Обновление соединения до WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Log.Error("Failed to upgrade to websocket",
			zap.Error(err))
		return
	}
	defer conn.Close()

	// Регистрация клиента
	h.service.RegisterClient(userID, conn)
	defer h.service.UnregisterClient(userID)

	for {
		var msg struct {
			Text string `json:"text"`
		}
		err := conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err) {
				logger.Log.Info("WebSocket closed",
					zap.String("user_id", userID),
					zap.Error(err))
			}
			break
		}

		// Обработка сообщения
		if err := h.service.HandleChatMessage(userID, msg.Text); err != nil {
			logger.Log.Error("Failed to handle chat message",
				zap.String("user_id", userID),
				zap.Error(err))
		}
	}
}
