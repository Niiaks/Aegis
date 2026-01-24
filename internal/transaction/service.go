package transaction

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Niiaks/Aegis/internal/middleware"
	"github.com/Niiaks/Aegis/internal/psp"
	"github.com/Niiaks/Aegis/internal/redis"
	"github.com/Niiaks/Aegis/pkg/types"
)

type TransactionService struct {
	repo           TransactionRepository
	redis          *redis.Client
	paystackClient *psp.PaystackClient
}

func NewTransactionService(repo TransactionRepository, redis *redis.Client, paystackClient *psp.PaystackClient) *TransactionService {
	return &TransactionService{
		repo:           repo,
		redis:          redis,
		paystackClient: paystackClient,
	}
}

func (ts *TransactionService) PaymentIntent(ctx context.Context, request *types.InitializePaymentRequest, idempotencyKey, requestID string) (*types.InitializePaymentResponse, error) {
	logger := middleware.GetLogger(ctx)

	logger.Info().Msg("Creating payment intent in service layer")

	//check idempotency
	cached, err := ts.redis.CheckAndSetIdempotency(ctx, idempotencyKey, 24*time.Hour)

	if cached != nil {
		logger.Info().Msg("Returning cached payment intent response due to idempotency key")
		var res types.InitializePaymentResponse
		json.Unmarshal(cached, &res)
		return &res, nil
	}

	if errors.Is(err, redis.ErrKeyExists) {
		logger.Warn().Msg("Request still in progress with same idempotency key")
		return nil, fmt.Errorf("request in progress: please retry later")
	}

	if err != nil {
		return nil, err
	}

	if !validateCurrency(request.Currency) {
		logger.Error().Msg("Unsupported currency")
		ts.redis.MarkIdempotencyFailed(ctx, idempotencyKey)
		return nil, fmt.Errorf("unsupported currency")
	}
	// check if amount is positive
	if request.Amount <= 0 {
		logger.Error().Msg("Amount must be more than zero")
		ts.redis.MarkIdempotencyFailed(ctx, idempotencyKey)
		return nil, fmt.Errorf("amount must be more than zero")
	}
	// additional checks can be added here

	// Call Paystack to initialize payment
	transactionID, err := ts.repo.PaymentIntent(ctx, request, idempotencyKey, requestID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create payment intent in repository layer")
		ts.redis.MarkIdempotencyFailed(ctx, idempotencyKey)
		return nil, fmt.Errorf("failed to create payment intent: %w", err)
	}

	paystackRes, err := ts.paystackClient.InitializePayment(ctx, request, transactionID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to initialize payment with Paystack")
		ts.redis.MarkIdempotencyFailed(ctx, idempotencyKey)
		return nil, fmt.Errorf("failed to initialize payment: %w", err)
	}

	// Cache the successful response for future duplicate requests
	responseBytes, _ := json.Marshal(paystackRes)
	ts.redis.MarkIdempotencyComplete(ctx, idempotencyKey, responseBytes, 24*time.Hour)

	return paystackRes, nil
}

func validateCurrency(currency string) bool {
	supportedCurrencies := map[string]bool{
		"USD": true,
		"EUR": true,
		"GHS": true,
	}
	return supportedCurrencies[currency]
}
