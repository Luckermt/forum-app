package repository

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/luckermt/forum-app/shared/pkg/config"
	"github.com/luckermt/forum-app/shared/pkg/logger"
	"github.com/luckermt/forum-app/shared/pkg/models"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(cfg config.PostgresConfig) (*PostgresRepository, error) {
	// Убедимся, что логгер инициализирован
	if logger.Log == nil {
		if err := logger.Init(); err != nil {
			return nil, fmt.Errorf("failed to initialize logger: %w", err)
		}
	}

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	logger.Log.Info("Connecting to PostgreSQL", zap.String("connStr",
		fmt.Sprintf("host=%s port=%s user=%s dbname=%s",
			cfg.Host, cfg.Port, cfg.User, cfg.DBName)))

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open PostgreSQL connection: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	logger.Log.Info("Successfully connected to PostgreSQL")
	return &PostgresRepository{db: db}, nil
}

func (r *PostgresRepository) CreateTopic(topic *models.Topic) error {
	query := `INSERT INTO topics (id, title, content, user_id, created_at, deleted) 
	          VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.Exec(query,
		topic.ID,
		topic.Title,
		topic.Content,
		topic.UserID,
		topic.CreatedAt,
		false, // deleted по умолчанию false
	)
	return err
}

func (r *PostgresRepository) GetTopics() ([]*models.Topic, error) {
	query := `SELECT id, title, content, user_id, created_at 
	          FROM topics WHERE deleted = false ORDER BY created_at DESC`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var topics []*models.Topic
	for rows.Next() {
		var topic models.Topic
		err := rows.Scan(
			&topic.ID,
			&topic.Title,
			&topic.Content,
			&topic.UserID,
			&topic.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		topics = append(topics, &topic)
	}

	return topics, nil
}

func (r *PostgresRepository) DeleteTopic(topicID string) error {
	query := `UPDATE topics SET deleted = true WHERE id = $1`
	_, err := r.db.Exec(query, topicID)
	return err
}

func (r *PostgresRepository) CreateMessage(message *models.Message) error {
	query := `INSERT INTO messages (id, topic_id, user_id, content, created_at, is_chat) 
	          VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.Exec(query,
		message.ID,
		message.TopicID,
		message.UserID,
		message.Content,
		message.CreatedAt,
		message.IsChat,
	)
	return err
}

func (r *PostgresRepository) GetMessagesByTopic(topicID string) ([]*models.Message, error) {
	query := `SELECT id, topic_id, user_id, content, created_at 
	          FROM messages WHERE topic_id = $1 AND is_chat = false 
	          ORDER BY created_at ASC`
	return r.queryMessages(query, topicID)
}

func (r *PostgresRepository) GetChatMessages() ([]*models.Message, error) {
	query := `SELECT id, topic_id, user_id, content, created_at 
	          FROM messages WHERE is_chat = true 
	          ORDER BY created_at ASC`
	return r.queryMessages(query)
}

func (r *PostgresRepository) queryMessages(query string, args ...interface{}) ([]*models.Message, error) {
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*models.Message
	for rows.Next() {
		var msg models.Message
		err := rows.Scan(
			&msg.ID,
			&msg.TopicID,
			&msg.UserID,
			&msg.Content,
			&msg.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, &msg)
	}

	return messages, nil
}

func (r *PostgresRepository) DeleteMessagesOlderThan(maxAge time.Duration) error {
	query := `DELETE FROM messages WHERE created_at < $1`
	_, err := r.db.Exec(query, time.Now().Add(-maxAge))
	return err
}

func (r *PostgresRepository) IsUserBlocked(userID string) (bool, error) {
	query := `SELECT blocked FROM users WHERE id = $1`
	var blocked bool
	err := r.db.QueryRow(query, userID).Scan(&blocked)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return blocked, err
}
func TestPostgresRepository_CreateTopic(t *testing.T) {
	cfg := config.PostgresConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "postgres",
		Password: "postgres",
		DBName:   "forum_test",
		SSLMode:  "disable",
	}

	repo, err := NewPostgresRepository(cfg)
	assert.NoError(t, err)

	topic := &models.Topic{
		ID:        "test-topic-id",
		Title:     "Test Topic",
		Content:   "Test Content",
		UserID:    "test-user-id",
		CreatedAt: time.Now(),
	}

	err = repo.CreateTopic(topic)
	assert.NoError(t, err)
}

func TestPostgresRepository_GetTopics(t *testing.T) {
	cfg := config.PostgresConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "postgres",
		Password: "postgres",
		DBName:   "forum_test",
		SSLMode:  "disable",
	}

	repo, err := NewPostgresRepository(cfg)
	assert.NoError(t, err)

	topics, err := repo.GetTopics()
	assert.NoError(t, err)
	assert.NotNil(t, topics)
}
