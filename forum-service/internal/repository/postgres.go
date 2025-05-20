package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/luckermt/forum-app/shared/pkg/config"
	"github.com/luckermt/forum-app/shared/pkg/logger"
	"github.com/luckermt/forum-app/shared/pkg/models"
	"go.uber.org/zap"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(cfg config.PostgresConfig) (*PostgresRepository, error) {
	if logger.Log == nil {
		if err := logger.Init(); err != nil {
			return nil, fmt.Errorf("failed to initialize logger: %w", err)
		}
	}

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	logger.Log.Info("Connecting to PostgreSQL",
		zap.String("host", cfg.Host),
		zap.String("dbname", cfg.DBName))

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
		false,
	)
	return err
}

func (r *PostgresRepository) GetTopics(page, limit int, search string) ([]*models.Topic, error) {
	offset := (page - 1) * limit
	query := `
		SELECT t.id, t.title, t.content, t.user_id, u.username, t.created_at
		FROM topics t
		JOIN users u ON t.user_id = u.id
		WHERE t.deleted = false
		AND ($1 = '' OR t.title ILIKE '%' || $1 || '%')
		ORDER BY t.created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(query, search, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query topics: %w", err)
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
			&topic.Username,
			&topic.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan topic: %w", err)
		}
		topics = append(topics, &topic)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return topics, nil
}

func (r *PostgresRepository) GetTopicsCount(search string) (int, error) {
	var count int
	query := `
		SELECT COUNT(*) 
		FROM topics
		WHERE deleted = false
		AND ($1 = '' OR title ILIKE '%' || $1 || '%')`

	err := r.db.QueryRow(query, search).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get topics count: %w", err)
	}
	return count, nil
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
	query := `
		SELECT m.id, m.topic_id, m.user_id, u.username, m.content, m.created_at
		FROM messages m
		JOIN users u ON m.user_id = u.id
		WHERE m.topic_id = $1 AND m.is_chat = false
		ORDER BY m.created_at ASC`

	rows, err := r.db.Query(query, topicID)
	if err != nil {
		return nil, fmt.Errorf("failed to query topic messages: %w", err)
	}
	defer rows.Close()

	var messages []*models.Message
	for rows.Next() {
		var msg models.Message
		err := rows.Scan(
			&msg.ID,
			&msg.TopicID,
			&msg.UserID,
			&msg.Username,
			&msg.Content,
			&msg.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, &msg)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return messages, nil
}

func (r *PostgresRepository) GetChatMessages() ([]*models.Message, error) {
	query := `
		SELECT m.id, m.topic_id, m.user_id, u.username, m.content, m.created_at
		FROM messages m
		JOIN users u ON m.user_id = u.id
		WHERE m.is_chat = true
		ORDER BY m.created_at ASC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query chat messages: %w", err)
	}
	defer rows.Close()

	var messages []*models.Message
	for rows.Next() {
		var msg models.Message
		err := rows.Scan(
			&msg.ID,
			&msg.TopicID,
			&msg.UserID,
			&msg.Username,
			&msg.Content,
			&msg.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, &msg)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return messages, nil
}

func (r *PostgresRepository) GetMessageCount(topicID string) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM messages WHERE topic_id = $1 AND is_chat = false`
	err := r.db.QueryRow(query, topicID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get message count: %w", err)
	}
	return count, nil
}

func (r *PostgresRepository) DeleteMessagesOlderThan(maxAge time.Duration) error {
	query := `DELETE FROM messages WHERE created_at < $1`
	_, err := r.db.Exec(query, time.Now().Add(-maxAge))
	return err
}

func (r *PostgresRepository) GetUsers(search string) ([]*models.User, error) {
	query := `
		SELECT id, username, email, role, blocked, created_at
		FROM users
		WHERE ($1 = '' OR username ILIKE '%' || $1 || '%' OR email ILIKE '%' || $1 || '%')
		ORDER BY created_at DESC`

	rows, err := r.db.Query(query, search)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.Role,
			&user.Blocked,
			&user.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, &user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return users, nil
}

func (r *PostgresRepository) BlockUser(userID string, blocked bool) error {
	query := `UPDATE users SET blocked = $1 WHERE id = $2`
	_, err := r.db.Exec(query, blocked, userID)
	return err
}

func (r *PostgresRepository) Close() error {
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}
