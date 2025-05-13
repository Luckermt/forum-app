package service

import (
	"time"

	"github.com/gorilla/websocket"
	"github.com/luckermt/forum-app/shared/pkg/models"
)

// ForumService определяет контракт для сервиса форума
type ForumService interface {
	// Topics
	CreateTopic(userID, title, content string) (*models.Topic, error)
	GetTopics() ([]*models.Topic, error)
	DeleteTopic(topicID, userID string) error

	// Messages
	CreateMessage(message *models.Message) error
	GetTopicMessages(topicID string) ([]*models.Message, error)
	GetChatMessages() ([]*models.Message, error)
	DeleteMessagesOlderThan(maxAge time.Duration) error

	// Chat
	RegisterClient(userID string, conn *websocket.Conn)
	UnregisterClient(userID string)
	HandleChatMessage(userID, text string) error
	CleanOldMessages(maxAge time.Duration)
	
	// Auth
	ValidateUser(token string) (string, error)
	IsUserAdmin(userID string) (bool, error)
}
type Repository interface {
	// Topics
	CreateTopic(topic *models.Topic) error
	GetTopics() ([]*models.Topic, error)
	DeleteTopic(topicID string) error

	// Messages
	CreateMessage(message *models.Message) error
	GetMessagesByTopic(topicID string) ([]*models.Message, error)
	GetChatMessages() ([]*models.Message, error)
	DeleteMessagesOlderThan(maxAge time.Duration) error
}
