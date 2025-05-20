package repository

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/luckermt/forum-app/shared/pkg/config"
	"github.com/luckermt/forum-app/shared/pkg/logger"
	"github.com/luckermt/forum-app/shared/pkg/models"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(cfg config.PostgresConfig) (*PostgresRepository, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Log.Info("Successfully connected to PostgreSQL")
	return &PostgresRepository{db: db}, nil
}

func (r *PostgresRepository) CreateUser(user *models.User) error {
	query := `INSERT INTO users (id, username, email, password, role, created_at, blocked) 
              VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.Exec(query,
		user.ID,
		user.Username,
		user.Email,
		user.Password,
		user.Role,
		user.CreatedAt,
		user.Blocked,
	)
	return err
}

func (r *PostgresRepository) GetUserByEmail(email string) (*models.User, error) {
	query := `SELECT id, username, email, password, role, created_at, blocked 
              FROM users WHERE email = $1`
	row := r.db.QueryRow(query, email)

	var user models.User
	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.CreatedAt,
		&user.Blocked,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *PostgresRepository) GetUserByID(userID string) (*models.User, error) {
	query := `SELECT id, username, email, password, role, created_at, blocked 
              FROM users WHERE id = $1`
	row := r.db.QueryRow(query, userID)

	var user models.User
	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.CreatedAt,
		&user.Blocked,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *PostgresRepository) BlockUser(userID string, blocked bool) error {
	query := `UPDATE users SET blocked = $1 WHERE id = $2`
	_, err := r.db.Exec(query, blocked, userID)
	if err != nil {
		return fmt.Errorf("failed to block user: %w", err)
	}
	return nil
}

func (r *PostgresRepository) UpdateUser(user *models.User) error {
	query := `UPDATE users SET username = $1, email = $2 WHERE id = $3`
	_, err := r.db.Exec(query, user.Username, user.Email, user.ID)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

func (r *PostgresRepository) Close() error {
	return r.db.Close()
}
