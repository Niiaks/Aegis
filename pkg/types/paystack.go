package types

import "time"

type PaystackWebhookEvent struct {
	Event string              `json:"event"`
	Data  PaystackWebhookData `json:"data"`
}
type PaystackWebhookData struct {
	ID              int64      `json:"id"`
	Domain          string     `json:"domain"`
	Status          string     `json:"status"`
	Reference       string     `json:"reference"`
	Amount          int64      `json:"amount"`
	Message         *string    `json:"message"`
	GatewayResponse string     `json:"gateway_response"`
	PaidAt          *time.Time `json:"paid_at"`
	CreatedAt       time.Time  `json:"created_at"`
	Channel         string     `json:"channel"`
	Currency        string     `json:"currency"`
	IPAddress       string     `json:"ip_address"`
	Metadata        struct {
		UserID        string `json:"user_id" validate:"required,uuid4"`
		TransactionID string `json:"transaction_id" validate:"required,uuid4"`
	} `json:"metadata"`
	Fees            int64                 `json:"fees"`
	Authorization   PaystackAuthorization `json:"authorization"`
	Customer        PaystackCustomer      `json:"customer"`
	RequestedAmount int64                 `json:"requested_amount"`
	Source          PaystackSource        `json:"source"`
}
type PaystackAuthorization struct {
	AuthorizationCode string `json:"authorization_code"`
	Bin               string `json:"bin"`
	Last4             string `json:"last4"`
	ExpMonth          string `json:"exp_month"`
	ExpYear           string `json:"exp_year"`
	Channel           string `json:"channel"`
	CardType          string `json:"card_type"`
	Bank              string `json:"bank"`
	CountryCode       string `json:"country_code"`
	Brand             string `json:"brand"`
	Reusable          bool   `json:"reusable"`
	Signature         string `json:"signature"`
}
type PaystackCustomer struct {
	ID           int64   `json:"id"`
	Email        string  `json:"email"`
	CustomerCode string  `json:"customer_code"`
	FirstName    *string `json:"first_name"`
	LastName     *string `json:"last_name"`
	Phone        *string `json:"phone"`
	RiskAction   string  `json:"risk_action"`
}
type PaystackSource struct {
	Type       string  `json:"type"`
	Source     string  `json:"source"`
	EntryPoint string  `json:"entry_point"`
	Identifier *string `json:"identifier"`
}

type BalanceUpdateEvent struct {
	TransactionID string `json:"transaction_id"`
	UserID        string `json:"user_id"`
	NetAmount     int64  `json:"net_amount"`
	Currency      string `json:"currency"`
}
