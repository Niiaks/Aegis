package wallet

import (
	"encoding/json"
	"net/http"

	"github.com/Niiaks/Aegis/internal/middleware"
	"github.com/Niiaks/Aegis/internal/model"
	"github.com/go-playground/validator/v10"
)

type WalletHandler struct {
	Service *WalletService
}

func NewWalletHandler(service *WalletService) *WalletHandler {
	return &WalletHandler{
		Service: service,
	}
}

var validate = validator.New()

func (wh *WalletHandler) CreateWallet(w http.ResponseWriter, r *http.Request) {
	var wallet model.Wallet
	ctx := r.Context()

	logger := middleware.GetLogger(ctx)

	logger.Info().Msg("Received request to create wallet")

	if err := json.NewDecoder(r.Body).Decode(&wallet); err != nil {
		logger.Error().Err(err).Msg("Failed to decode wallet creation request")
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if err := validate.Struct(&wallet); err != nil {
		logger.Error().Err(err).Msg("Validation error on wallet creation request")
		http.Error(w, "Validation error: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := wh.Service.CreateWallet(ctx, &wallet); err != nil {
		logger.Error().Err(err).Msg("Failed to create wallet")
		http.Error(w, "Failed to create wallet", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"message":   "Wallet created successfully",
		"wallet_id": wallet.ID,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	logger.Info().Msgf("Wallet created successfully with ID: %s", wallet.ID)
}
