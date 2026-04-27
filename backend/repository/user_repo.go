package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/irham/topup-backend/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("not found")

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, u *model.User) error {
	err := r.db.QueryRow(ctx, `
		INSERT INTO users (email, phone, username, password_hash, full_name, role, status)
		VALUES ($1, NULLIF($2, ''), $3, $4, NULLIF($5, ''), 'buyer', 'unverified')
		RETURNING id, created_at
	`, u.Email, u.Phone, u.Username, u.PasswordHash, u.FullName).Scan(&u.ID, &u.CreatedAt)
	if err != nil {
		return fmt.Errorf("user create: %w", err)
	}
	u.Role = "buyer"
	u.Status = "unverified"
	return nil
}

func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var u model.User
	err := r.db.QueryRow(ctx, `
		SELECT id, email, COALESCE(phone, ''), username, password_hash,
		       COALESCE(full_name, ''), COALESCE(avatar_url, ''),
		       role::text, status::text, last_login_at, created_at
		FROM users WHERE email = $1
	`, email).Scan(
		&u.ID, &u.Email, &u.Phone, &u.Username, &u.PasswordHash,
		&u.FullName, &u.AvatarURL, &u.Role, &u.Status, &u.LastLoginAt, &u.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("user find by email: %w", err)
	}
	return &u, nil
}

func (r *UserRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var u model.User
	err := r.db.QueryRow(ctx, `
		SELECT id, email, COALESCE(phone, ''), username, password_hash,
		       COALESCE(full_name, ''), COALESCE(avatar_url, ''),
		       role::text, status::text, last_login_at, created_at
		FROM users WHERE id = $1
	`, id).Scan(
		&u.ID, &u.Email, &u.Phone, &u.Username, &u.PasswordHash,
		&u.FullName, &u.AvatarURL, &u.Role, &u.Status, &u.LastLoginAt, &u.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("user find by id: %w", err)
	}
	return &u, nil
}

func (r *UserRepo) TouchLastLogin(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE users SET last_login_at = NOW() WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("touch last login: %w", err)
	}
	return nil
}
