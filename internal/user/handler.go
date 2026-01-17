package user

import (
	"encoding/json"
	"net/http"

	"github.com/Niiaks/Aegis/internal/middleware"
	"github.com/Niiaks/Aegis/internal/model"
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

func (uh *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var user model.User
	ctx := r.Context()

	logger := middleware.GetLogger(ctx)
	logger.Info().Msg("Received request to create user")
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if err := validate.Struct(&user); err != nil {
		http.Error(w, "Validation error: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := uh.service.CreateUser(ctx, &user); err != nil {
		logger.Error().Err(err).Msg("Failed to create user")
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}
