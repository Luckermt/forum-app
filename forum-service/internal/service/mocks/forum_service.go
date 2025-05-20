package mocks

import (
	"time"

	"github.com/gorilla/websocket"
	"github.com/luckermt/forum-app/shared/pkg/models"
	"github.com/stretchr/testify/mock"
)

type ForumService struct {
	mock.Mock
}

func (m *ForumService) CreateTopic(userID, title, content string) (*models.Topic, error) {
	args := m.Called(userID, title, content)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Topic), args.Error(1)
}
func (m *ForumService) BlockUser(userID string, blocked bool) error {
	args := m.Called(userID, blocked)
	return args.Error(0)
}
func (m *ForumService) GetTopics(page, limit int, search string) ([]*models.Topic, int, error) {
	args := m.Called(page, limit, search)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*models.Topic), args.Int(1), args.Error(2)
}

func (m *ForumService) DeleteTopic(topicID, userID string) error {
	args := m.Called(topicID, userID)
	return args.Error(0)
}

func (m *ForumService) CreateMessage(message *models.Message) error {
	args := m.Called(message)
	return args.Error(0)
}

func (m *ForumService) GetTopicMessages(topicID string) ([]*models.Message, error) {
	args := m.Called(topicID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Message), args.Error(1)
}

func (m *ForumService) GetChatMessages() ([]*models.Message, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Message), args.Error(1)
}

func (m *ForumService) DeleteMessagesOlderThan(maxAge time.Duration) error {
	args := m.Called(maxAge)
	return args.Error(0)
}

func (m *ForumService) RegisterClient(userID string, conn *websocket.Conn) {
	m.Called(userID, conn)
}

func (m *ForumService) UnregisterClient(userID string) {
	m.Called(userID)
}

func (m *ForumService) HandleChatMessage(userID, text string) error {
	args := m.Called(userID, text)
	return args.Error(0)
}

func (m *ForumService) CleanOldMessages(maxAge time.Duration) {
	m.Called(maxAge)
}

func (m *ForumService) ValidateUser(token string) (string, error) {
	args := m.Called(token)
	return args.String(0), args.Error(1)
}

func (m *ForumService) IsUserAdmin(userID string) (bool, error) {
	args := m.Called(userID)
	return args.Bool(0), args.Error(1)
}
func (m *ForumService) GetUsers(search string) ([]*models.User, error) {
	args := m.Called(search)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}
