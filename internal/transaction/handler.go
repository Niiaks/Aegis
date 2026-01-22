package transaction

import (
	"encoding/json"
	"net/http"

	"github.com/Niiaks/Aegis/internal/middleware"
	"github.com/Niiaks/Aegis/pkg/types"
	"github.com/go-playground/validator/v10"
)

type TransactionHandler struct {
	transactionService *TransactionService
}

func NewTransactionHandler(transactionService *TransactionService) *TransactionHandler {
	return &TransactionHandler{
		transactionService: transactionService,
	}
}

var validate = validator.New()

func (th *TransactionHandler) PaymentIntent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	logger := middleware.GetLogger(ctx)
	logger.Info().Msg("Received request to create payment intent")

	//get idempotency key from header
	idemKey := r.Header.Get("Idempotency-Key")
	if idemKey == "" {
		logger.Error().Msg("Idempotency-Key header is missing")
		http.Error(w, "Idempotency-Key header is required", http.StatusBadRequest)
		return
	}
	var req types.InitializePaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error().Err(err).Msg("Failed to decode payment intent request")
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if err := validate.Struct(&req); err != nil {
		logger.Error().Err(err).Msg("Validation error on payment intent request")
		http.Error(w, "Validation error: "+err.Error(), http.StatusBadRequest)
		return
	}

	res, err := th.transactionService.PaymentIntent(ctx, &req, idemKey)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create payment intent")
		http.Error(w, "Failed to create payment intent: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
	logger.Info().Msg("Payment intent created successfully")
}
