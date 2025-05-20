package service

import (
	"time"

	"github.com/gorilla/websocket"
	"github.com/luckermt/forum-app/shared/pkg/models"
)

type ForumService interface {
	// Topics
	CreateTopic(userID, title, content string) (*models.Topic, error)
	GetTopics(page, limit int, search string) ([]*models.Topic, int, error)
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

	// Admin
	GetUsers(search string) ([]*models.User, error)
	BlockUser(userID string, blocked bool) error
}

type Repository interface {
	// Topics
	CreateTopic(topic *models.Topic) error
	GetTopics(page, limit int, search string) ([]*models.Topic, error)
	GetTopicsCount(search string) (int, error)
	DeleteTopic(topicID string) error

	// Messages
	CreateMessage(message *models.Message) error
	GetMessagesByTopic(topicID string) ([]*models.Message, error)
	GetChatMessages() ([]*models.Message, error)
	GetMessageCount(topicID string) (int, error)
	DeleteMessagesOlderThan(maxAge time.Duration) error

	// Users
	GetUsers(search string) ([]*models.User, error)
	BlockUser(userID string, blocked bool) error
}

type AuthClient interface {
	ValidateToken(token string) (string, error)
	IsUserAdmin(userID string) (bool, error)
	GetUserInfo(userID string) (*models.User, error)
}
