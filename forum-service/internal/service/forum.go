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
	onlineUsers   map[string]bool
	onlineMutex   sync.Mutex
}

func NewForumService(repo Repository, authClient AuthClient) *forumServiceImpl {
	service := &forumServiceImpl{
		repo:          repo,
		authClient:    authClient,
		chatClients:   make(map[string]*websocket.Conn),
		broadcastChan: make(chan models.Message, 100),
		onlineUsers:   make(map[string]bool),
	}
	go service.startMessageBroadcaster()
	go service.cleanupInactiveUsers()
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

	// Создаем первое сообщение в теме
	message := &models.Message{
		ID:        generateID(),
		TopicID:   topic.ID,
		UserID:    userID,
		Content:   content,
		CreatedAt: time.Now(),
		IsChat:    false,
	}

	if err := s.repo.CreateMessage(message); err != nil {
		logger.Log.Error("Failed to create initial topic message",
			zap.String("topic_id", topic.ID),
			zap.Error(err))
		return nil, err
	}

	return topic, nil
}

func (s *forumServiceImpl) GetTopics(page, limit int, search string) ([]*models.Topic, int, error) {
	topics, err := s.repo.GetTopics(page, limit, search)
	if err != nil {
		logger.Log.Error("Failed to get topics", 
			zap.Error(err),
			zap.Int("page", page),
			zap.Int("limit", limit))
		return nil, 0, err
	}

	total, err := s.repo.GetTopicsCount(search)
	if err != nil {
		logger.Log.Error("Failed to get topics count",
			zap.Error(err))
		return nil, 0, err
	}

	// Получаем информацию о пользователях для тем
	for _, topic := range topics {
		user, err := s.authClient.GetUserInfo(topic.UserID)
		if err != nil {
			logger.Log.Warn("Failed to get user info for topic",
				zap.String("user_id", topic.UserID),
				zap.Error(err))
			continue
		}
		topic.Username = user.Username
		
		// Получаем количество сообщений в теме
		count, err := s.repo.GetMessageCount(topic.ID)
		if err != nil {
			logger.Log.Warn("Failed to get message count for topic",
				zap.String("topic_id", topic.ID),
				zap.Error(err))
			continue
		}
		topic.MessageCount = count
	}

	return topics, total, nil
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
	// Получаем информацию о пользователе
	user, err := s.authClient.GetUserInfo(message.UserID)
	if err != nil {
		logger.Log.Error("Failed to get user info for message",
			zap.String("user_id", message.UserID),
			zap.Error(err))
		return err
	}
	message.Username = user.Username

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

	// Добавляем информацию о пользователях
	for _, msg := range messages {
		user, err := s.authClient.GetUserInfo(msg.UserID)
		if err != nil {
			logger.Log.Warn("Failed to get user info for message",
				zap.String("user_id", msg.UserID),
				zap.Error(err))
			continue
		}
		msg.Username = user.Username
	}

	return messages, nil
}

func (s *forumServiceImpl) GetChatMessages() ([]*models.Message, error) {
	messages, err := s.repo.GetChatMessages()
	if err != nil {
		logger.Log.Error("Failed to get chat messages", zap.Error(err))
		return nil, err
	}

	// Добавляем информацию о пользователях
	for _, msg := range messages {
		user, err := s.authClient.GetUserInfo(msg.UserID)
		if err != nil {
			logger.Log.Warn("Failed to get user info for chat message",
				zap.String("user_id", msg.UserID),
				zap.Error(err))
			continue
		}
		msg.Username = user.Username
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
	
	s.onlineMutex.Lock()
	s.onlineUsers[userID] = true
	s.onlineMutex.Unlock()
	
	// Отправляем обновленное количество онлайн пользователей
	s.broadcastOnlineCount()
	
	logger.Log.Info("New chat client registered",
		zap.String("user_id", userID))
}

func (s *forumServiceImpl) UnregisterClient(userID string) {
	s.clientsMutex.Lock()
	defer s.clientsMutex.Unlock()
	
	delete(s.chatClients, userID)
	
	s.onlineMutex.Lock()
	delete(s.onlineUsers, userID)
	s.onlineMutex.Unlock()
	
	// Отправляем обновленное количество онлайн пользователей
	s.broadcastOnlineCount()
	
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

func (s *forumServiceImpl) cleanupInactiveUsers() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.clientsMutex.Lock()
		for userID, conn := range s.chatClients {
			if err := conn.WriteJSON(map[string]string{"type": "ping"}); err != nil {
				logger.Log.Info("Removing inactive connection",
					zap.String("user_id", userID),
					zap.Error(err))
				delete(s.chatClients, userID)
				
				s.onlineMutex.Lock()
				delete(s.onlineUsers, userID)
				s.onlineMutex.Unlock()
			}
		}
		s.clientsMutex.Unlock()
		
		// Отправляем обновленное количество онлайн пользователей
		s.broadcastOnlineCount()
	}
}

func (s *forumServiceImpl) broadcastOnlineCount() {
	s.onlineMutex.Lock()
	count := len(s.onlineUsers)
	s.onlineMutex.Unlock()

	message := struct {
		Type  string `json:"type"`
		Count int    `json:"count"`
	}{
		Type:  "online_count",
		Count: count,
	}

	s.clientsMutex.Lock()
	defer s.clientsMutex.Unlock()

	for userID, conn := range s.chatClients {
		if err := conn.WriteJSON(message); err != nil {
			logger.Log.Error("Failed to send online count",
				zap.String("user_id", userID),
				zap.Error(err))
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

// Admin methods
func (s *forumServiceImpl) GetUsers(search string) ([]*models.User, error) {
	return s.repo.GetUsers(search)
}

func (s *forumServiceImpl) BlockUser(userID string, blocked bool) error {
	return s.repo.BlockUser(userID, blocked)
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