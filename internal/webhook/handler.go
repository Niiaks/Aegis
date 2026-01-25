package webhook

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"

	"github.com/Niiaks/Aegis/internal/kafka"
	"github.com/Niiaks/Aegis/internal/middleware"
	"github.com/Niiaks/Aegis/pkg/types"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WebhookHandler struct {
	env         string
	kafkaClient *kafka.Producer
	db          *pgxpool.Pool
}

func NewWebhookHandler(env string, kafkaClient *kafka.Producer, db *pgxpool.Pool) *WebhookHandler {
	return &WebhookHandler{
		env:         env,
		kafkaClient: kafkaClient,
		db:          db,
	}
}

// VerifyPaystackSignature validates the webhook came from Paystack.
// signature: value from "x-paystack-signature" header
// payload: raw request body bytes
// secretKey: your Paystack secret key from config
func verifyWebhookSignature(payload []byte, signature, secret string) bool {

	mac := hmac.New(sha512.New, []byte(secret))
	_, err := mac.Write(payload)
	if err != nil {
		return false
	}

	expectedSig := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(expectedSig), []byte(signature))
}

func (h *WebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := middleware.GetLogger(ctx)
	header := r.Header.Get("x-paystack-signature")
	if header == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	logger.Info().Msg("Received webhook request")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to read request body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	logger.Info().Msg("Verifying webhook signature")
	if !verifyWebhookSignature(body, header, h.env) {
		logger.Error().Msg("Invalid webhook signature")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	logger.Info().Msg("Webhook signature verified")

	w.WriteHeader(http.StatusOK)
	var event types.PaystackWebhookEvent
	json.Unmarshal(body, &event)
	if event.Event == "charge.success" {
		// Store in outbox for reliable delivery
		_, err = h.db.Exec(ctx, `
			INSERT INTO transaction_outbox (event_type, payload, partition_key, status)
			VALUES ($1, $2, $3, $4)
		`, kafka.EventWebhookReceived, body, event.Data.Metadata.UserID, "pending")

		if err != nil {
			logger.Error().Err(err).Msg("Failed to store webhook in outbox")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		logger.Info().Str("event", event.Event).Str("user_id", event.Data.Metadata.UserID).Msg("Webhook stored in outbox")
	}
}
