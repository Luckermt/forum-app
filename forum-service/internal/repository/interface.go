package repository

import (
	"time"

	"github.com/luckermt/forum-app/shared/pkg/models"
)

type ForumRepository interface {
	// Topics
	CreateTopic(topic *models.Topic) error
	GetTopics() ([]*models.Topic, error)
	DeleteTopic(topicID string) error

	// Messages
	CreateMessage(message *models.Message) error
	GetMessagesByTopic(topicID string) ([]*models.Message, error)
	GetChatMessages() ([]*models.Message, error)
	DeleteMessagesOlderThan(t time.Duration) error

	// Moderation
	IsUserBlocked(userID string) (bool, error)
}
