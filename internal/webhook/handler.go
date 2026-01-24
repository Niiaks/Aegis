package webhook

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"io"
	"net/http"

	"github.com/Niiaks/Aegis/internal/middleware"
)

type WebhookHandler struct {
	env string
}

func NewWebhookHandler(env string) *WebhookHandler {
	return &WebhookHandler{
		env: env,
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
	logger.Info().Msg(string(body))
	w.WriteHeader(http.StatusOK)
	//send to kafka here TODO but check event data to send to right kafka consumer

}
