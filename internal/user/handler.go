package user

import (
	"github.com/go-playground/validator/v10"
)

type UserHandler struct {
	service *UserService
}

func NewUserHandler(service *UserService) *UserHandler {
	return &UserHandler{
		service: service,
	}
}

var validate = validator.New()
