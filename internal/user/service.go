package user

import (
	"context"

	"github.com/Niiaks/Aegis/internal/model"
)

type UserService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (us *UserService) CreateUser(ctx context.Context, user *model.User) error {
	return us.repo.CreateUser(ctx, user)
}
