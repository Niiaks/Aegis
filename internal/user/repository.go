package user

import (
	"context"
	"fmt"
	"time"

	"github.com/Niiaks/Aegis/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *model.User) error
}

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{db: db}
}

func (ur *UserRepo) CreateUser(ctx context.Context, user *model.User) error {
	err := ur.db.QueryRow(ctx, "INSERT INTO users (name, email, platform_id, psp_id, created_at, updated_at) VALUES ($1, $2,$3,$4,$5,$6) RETURNING id", user.Name, user.Email, user.PlatformID, user.PspID, time.Now(), time.Now()).Scan(&user.ID)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}
