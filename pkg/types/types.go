package types

type InitializePaymentRequest struct {
	Email       string `json:"email" validate:"required,email"`
	CallbackURL string `json:"callback_url,omitempty"`
	Amount      int64  `json:"amount" validate:"required,gte=0"`
	Currency    string `json:"currency" validate:"required,len=3"`
	UserID      string `json:"user_id" validate:"required,uuid4"`
	Status      string `json:"status" validate:"required,oneof=pending"`
	Type        string `json:"type" validate:"required,oneof=payment_intent"`
}

type InitializePaymentResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		AuthorizationURL string `json:"authorization_url"`
		AccessCode       string `json:"access_code"`
		Reference        string `json:"reference"`
	} `json:"data"`
}
