package transaction

import (
	"context"

	"github.com/Niiaks/Aegis/pkg/types"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TransactionRepository interface {
	PaymentIntent(ctx context.Context, request *types.InitializePaymentRequest, idempotencyKey string) error
}

type TransactionRepo struct {
	db *pgxpool.Pool
}

func NewTransactionRepository(db *pgxpool.Pool) *TransactionRepo {
	return &TransactionRepo{
		db: db,
	}
}

func (tr *TransactionRepo) PaymentIntent(ctx context.Context, request *types.InitializePaymentRequest, idempotencyKey string) error {
	sql := `INSERT INTO transactions (user_id, idempotency_key, amount, currency, status, type) VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := tr.db.Exec(ctx, sql,
		request.UserID,
		idempotencyKey,
		request.Amount,
		request.Currency,
		request.Status,
		request.Type,
	)
	return err
}
