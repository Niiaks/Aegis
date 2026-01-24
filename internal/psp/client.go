package psp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Niiaks/Aegis/pkg/types"
	"github.com/rs/zerolog/log"
)

type PaystackClient struct {
	httpClient *http.Client
	secretKey  string
	baseURL    string
}

type Client interface {
	InitializePayment(ctx context.Context)
	CreateTransfer(ctx context.Context)
}

func NewPaystackClient(secretKey string) *PaystackClient {
	return &PaystackClient{
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 50,
				IdleConnTimeout:     90 * time.Second,
				DisableKeepAlives:   false,
			},
		},
		secretKey: secretKey,
		baseURL:   "https://api.paystack.co",
	}
}

func (c *PaystackClient) InitializePayment(ctx context.Context, req *types.InitializePaymentRequest, transactionID string) (*types.InitializePaymentResponse, error) {
	req.Metadata.TransactionID = transactionID
	respBody, err := c.doRequest(ctx, http.MethodPost, "/transaction/initialize", req)
	if err != nil {
		return nil, err
	}

	var resp types.InitializePaymentResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !resp.Status {
		return nil, fmt.Errorf("paystack error: %s", resp.Message)
	}

	return &resp, nil
}

func (c *PaystackClient) doRequest(ctx context.Context, method, path string, body any) ([]byte, error) {
	url := c.baseURL + path

	var reqBody io.Reader
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			log.Error().Err(err).Msg("Failed to marshal request body")
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}
		reqBody = bytes.NewReader(jsonBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create HTTP request")
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.secretKey)
	req.Header.Set("Content-Type", "application/json")

	start := time.Now()
	resp, err := c.httpClient.Do(req)
	duration := time.Since(start).Milliseconds()
	if err != nil {
		log.Error().Err(err).
			Str("method", method).
			Str("url", url).
			Int64("duration_ms", duration).
			Msg("HTTP request failed")
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).
			Str("method", method).
			Str("url", url).
			Int64("duration_ms", duration).
			Msg("Failed to read response body")
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		log.Error().
			Int("status", resp.StatusCode).
			Str("method", method).
			Str("url", url).
			Int64("duration_ms", duration).
			Str("body", string(respBody)).
			Msg("Paystack API error response")
		return nil, fmt.Errorf("paystack error: status=%d body=%s", resp.StatusCode, string(respBody))
	}

	log.Info().
		Int("status", resp.StatusCode).
		Str("method", method).
		Str("url", url).
		Int64("duration_ms", duration).
		Msg("Paystack API request successful")

	return respBody, nil
}
