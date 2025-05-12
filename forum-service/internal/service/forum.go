package service

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/luckermt/shared/pkg/logger"
	"github.com/luckermt/shared/pkg/models"
	"go.uber.org/zap"
)

type ForumService struct {
	repo       Repository
	authClient AuthClient
	clients    map[string]*websocket.Conn
	broadcast  chan models.Message
}

func NewForumService(repo Repository, authClient AuthClient) *ForumService {
	service := &ForumService{
		repo:       repo,
		authClient: authClient,
		clients:    make(map[string]*websocket.Conn),
		broadcast:  make(chan models.Message),
	}

	go service.handleBroadcast()
	return service
}

func (s *ForumService) CreateTopic(userID, title, content string) (*models.Topic, error) {
	topic := &models.Topic{
		ID:        uuid.New().String(),
		Title:     title,
		Content:   content,
		UserID:    userID,
		CreatedAt: time.Now(),
	}

	err := s.repo.CreateTopic(topic)
	if err != nil {
		return nil, err
	}

	return topic, nil
}

func (s *ForumService) GetTopics() ([]*models.Topic, error) {
	return s.repo.GetTopics()
}

func (s *ForumService) DeleteTopic(topicID, userID string) error {
	// Проверка прав администратора
	isAdmin, err := s.isAdmin(userID)
	if err != nil {
		return err
	}

	if !isAdmin {
		return errors.New("permission denied")
	}

	return s.repo.DeleteTopic(topicID)
}

func (s *ForumService) handleBroadcast() {
	for msg := range s.broadcast {
		for _, client := range s.clients {
			err := client.WriteJSON(msg)
			if err != nil {
				logger.Log.Error("Failed to send message", zap.Error(err))
				client.Close()
				delete(s.clients, msg.UserID)
			}
		}
	}
}

func (s *ForumService) CleanOldMessages(maxAge time.Duration) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		err := s.repo.DeleteMessagesOlderThan(time.Now().Add(-maxAge))
		if err != nil {
			logger.Log.Error("Failed to clean old messages", zap.Error(err))
		}
	}
}
