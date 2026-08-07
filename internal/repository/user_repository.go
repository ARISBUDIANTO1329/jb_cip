package repository

import (
	"database/sql"
	"fmt"

	"github.com/jaybani/jb_cip/internal/domain"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindByEmail(email string) (*domain.User, error) {
	query := `
		SELECT id, email, password_hash, name, COALESCE(avatar_url, ''), status,
		       email_verified_at, last_login_at, created_at, updated_at
		FROM core.users
		WHERE email = $1 AND deleted_at IS NULL
	`

	user := &domain.User{}
	err := r.db.QueryRow(query, email).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Name,
		&user.AvatarURL, &user.Status, &user.EmailVerifiedAt,
		&user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) FindByID(id string) (*domain.User, error) {
	query := `
		SELECT id, email, password_hash, name, COALESCE(avatar_url, ''), status,
		       email_verified_at, last_login_at, created_at, updated_at
		FROM core.users
		WHERE id = $1 AND deleted_at IS NULL
	`

	user := &domain.User{}
	err := r.db.QueryRow(query, id).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Name,
		&user.AvatarURL, &user.Status, &user.EmailVerifiedAt,
		&user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) UpdateLastLogin(id string) error {
	query := `UPDATE core.users SET last_login_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *UserRepository) Create(user *domain.User) error {
	query := `
		INSERT INTO core.users (email, password_hash, name, status, email_verified_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`

	return r.db.QueryRow(
		query,
		user.Email, user.PasswordHash, user.Name, user.Status, user.EmailVerifiedAt,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}
