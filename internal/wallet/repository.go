package wallet

import (
	"context"

	"github.com/Niiaks/Aegis/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WalletRepository interface {
	CreateWallet(ctx context.Context, wallet *model.Wallet) error
}

type WalletRepo struct {
	db *pgxpool.Pool
}

func NewWalletRepository(db *pgxpool.Pool) *WalletRepo {
	return &WalletRepo{db: db}
}

func (wr *WalletRepo) CreateWallet(ctx context.Context, wallet *model.Wallet) error {
	err := wr.db.QueryRow(ctx, "INSERT INTO wallets (user_id, currency,type) VALUES ($1, $2, $3) RETURNING id", wallet.UserID, wallet.Currency, wallet.Type).Scan(&wallet.ID)
	if err != nil {
		return err
	}
	return nil
}
