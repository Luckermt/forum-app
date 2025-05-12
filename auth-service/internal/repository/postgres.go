package repository

import (
	"database/sql"
	"fmt"

	"github.com/luckermt/shared/pkg/config"
	"github.com/luckermt/shared/pkg/logger"
	"github.com/luckermt/shared/pkg/models"
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
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresRepository{db: db}, nil
}

func (r *PostgresRepository) CreateUser(user *models.User) error {
	query := `INSERT INTO users (id, username, email, password, role, created_at, blocked) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.Exec(query, user.ID, user.Username, user.Email, user.Password, user.Role, user.CreatedAt, user.Blocked)
	return err
}

func (r *PostgresRepository) GetUserByEmail(email string) (*models.User, error) {
	query := `SELECT id, username, email, password, role, created_at, blocked FROM users WHERE email = $1`
	row := r.db.QueryRow(query, email)

	var user models.User
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.Role, &user.CreatedAt, &user.Blocked)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *PostgresRepository) BlockUser(userID string) error {
	query := `UPDATE users SET blocked = true WHERE id = $1`
	_, err := r.db.Exec(query, userID)
	return err
}