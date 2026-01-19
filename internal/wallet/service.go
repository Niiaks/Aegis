package wallet

import (
	"context"

	"github.com/Niiaks/Aegis/internal/middleware"
	"github.com/Niiaks/Aegis/internal/model"
)

type WalletService struct {
	walletRepo WalletRepository
}

func NewWalletService(walletRepo WalletRepository) *WalletService {
	return &WalletService{
		walletRepo: walletRepo,
	}
}

func (ws *WalletService) CreateWallet(ctx context.Context, wallet *model.Wallet) error {
	logger := middleware.GetLogger(ctx)
	logger.Info().Msg("Creating wallet in service layer")
	return ws.walletRepo.CreateWallet(ctx, wallet)
}
