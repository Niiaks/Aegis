package wallet

import (
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
