package service

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/luckermt/forum-app/shared/pkg/logger"
	"github.com/luckermt/forum-app/shared/pkg/models"
	"go.uber.org/zap"
)

type forumServiceImpl struct {
	repo          Repository
	authClient    AuthClient
	chatClients   map[string]*websocket.Conn
	clientsMutex  sync.Mutex
	broadcastChan chan models.Message
}

func NewForumService(repo Repository, authClient AuthClient) *forumServiceImpl {
	service := &forumServiceImpl{
		repo:          repo,
		authClient:    authClient,
		chatClients:   make(map[string]*websocket.Conn),
		broadcastChan: make(chan models.Message, 100),
	}
	go service.startMessageBroadcaster()
	return service
}

// Topic methods
func (s *forumServiceImpl) CreateTopic(userID, title, content string) (*models.Topic, error) {
	topic := &models.Topic{
		ID:        generateID(),
		Title:     title,
		Content:   content,
		UserID:    userID,
		CreatedAt: time.Now(),
	}

	if err := s.repo.CreateTopic(topic); err != nil {
		logger.Log.Error("Failed to create topic",
			zap.String("user_id", userID),
			zap.Error(err))
		return nil, err
	}

	return topic, nil
}

func (s *forumServiceImpl) GetTopics() ([]*models.Topic, error) {
	topics, err := s.repo.GetTopics()
	if err != nil {
		logger.Log.Error("Failed to get topics", zap.Error(err))
		return nil, err
	}
	return topics, nil
}

func (s *forumServiceImpl) DeleteTopic(topicID, userID string) error {
	isAdmin, err := s.authClient.IsUserAdmin(userID)
	if err != nil {
		logger.Log.Error("Failed to check admin status",
			zap.String("user_id", userID),
			zap.Error(err))
		return err
	}

	if !isAdmin {
		return ErrForbidden
	}

	return s.repo.DeleteTopic(topicID)
}

// Message methods
func (s *forumServiceImpl) CreateMessage(message *models.Message) error {
	if err := s.repo.CreateMessage(message); err != nil {
		logger.Log.Error("Failed to create message",
			zap.String("user_id", message.UserID),
			zap.Error(err))
		return err
	}

	if message.IsChat {
		s.broadcastChan <- *message
	}

	return nil
}

func (s *forumServiceImpl) GetTopicMessages(topicID string) ([]*models.Message, error) {
	messages, err := s.repo.GetMessagesByTopic(topicID)
	if err != nil {
		logger.Log.Error("Failed to get topic messages",
			zap.String("topic_id", topicID),
			zap.Error(err))
		return nil, err
	}
	return messages, nil
}

func (s *forumServiceImpl) GetChatMessages() ([]*models.Message, error) {
	messages, err := s.repo.GetChatMessages()
	if err != nil {
		logger.Log.Error("Failed to get chat messages", zap.Error(err))
		return nil, err
	}
	return messages, nil
}

func (s *forumServiceImpl) DeleteMessagesOlderThan(maxAge time.Duration) error {
	if err := s.repo.DeleteMessagesOlderThan(maxAge); err != nil {
		logger.Log.Error("Failed to delete old messages",
			zap.Duration("max_age", maxAge),
			zap.Error(err))
		return err
	}
	return nil
}

// Chat methods
func (s *forumServiceImpl) RegisterClient(userID string, conn *websocket.Conn) {
	s.clientsMutex.Lock()
	defer s.clientsMutex.Unlock()
	s.chatClients[userID] = conn
	logger.Log.Info("New chat client registered",
		zap.String("user_id", userID))
}

func (s *forumServiceImpl) UnregisterClient(userID string) {
	s.clientsMutex.Lock()
	defer s.clientsMutex.Unlock()
	delete(s.chatClients, userID)
	logger.Log.Info("Chat client unregistered",
		zap.String("user_id", userID))
}

func (s *forumServiceImpl) HandleChatMessage(userID, text string) error {
	msg := models.Message{
		ID:        generateID(),
		UserID:    userID,
		Content:   text,
		CreatedAt: time.Now(),
		IsChat:    true,
	}

	return s.CreateMessage(&msg)
}

func (s *forumServiceImpl) CleanOldMessages(maxAge time.Duration) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		if err := s.DeleteMessagesOlderThan(maxAge); err != nil {
			logger.Log.Error("Failed to clean old messages", zap.Error(err))
		}
	}
}

// Auth methods
func (s *forumServiceImpl) ValidateUser(token string) (string, error) {
	return s.authClient.ValidateToken(token)
}

func (s *forumServiceImpl) IsUserAdmin(userID string) (bool, error) {
	return s.authClient.IsUserAdmin(userID)
}

// Internal methods
func (s *forumServiceImpl) startMessageBroadcaster() {
	for msg := range s.broadcastChan {
		s.clientsMutex.Lock()
		for userID, conn := range s.chatClients {
			if err := conn.WriteJSON(msg); err != nil {
				logger.Log.Error("Failed to send message",
					zap.String("user_id", userID),
					zap.Error(err))
				s.UnregisterClient(userID)
			}
		}
		s.clientsMutex.Unlock()
	}
}

func generateID() string {
	return uuid.New().String()
}
