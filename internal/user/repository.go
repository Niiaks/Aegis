package user

import (
	"github.com/Niiaks/Aegis/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	CreateUser(user *model.User) error
	GetUserByID(id string) (*model.User, error)
}

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{db: db}
}
