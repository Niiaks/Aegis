package transaction

import (
	"context"
	"encoding/json"

	"github.com/Niiaks/Aegis/pkg/types"
	"github.com/jackc/pgx/v5/pgxpool"
)

var Event string

var PaymentIntentEvent = "aegis.aegis.payment.created"

type TransactionRepository interface {
	PaymentIntent(ctx context.Context, request *types.InitializePaymentRequest, idempotencyKey, correlationID string) error
}

type TransactionRepo struct {
	db *pgxpool.Pool
}

func NewTransactionRepository(db *pgxpool.Pool) *TransactionRepo {
	return &TransactionRepo{
		db: db,
	}
}

func (tr *TransactionRepo) PaymentIntent(ctx context.Context, request *types.InitializePaymentRequest, idempotencyKey, correlationID string) error {
	tx, err := tr.db.Begin(ctx)
	if err != nil {
		return err
	}

	transactionQuery := `INSERT INTO transactions (user_id, idempotency_key, amount, currency, status, type) VALUES ($1, $2, $3, $4, $5, $6)`

	_, err = tx.Exec(ctx, transactionQuery,
		request.UserID,
		idempotencyKey,
		request.Amount,
		request.Currency,
		request.Status,
		request.Type,
	)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	payload, err := json.Marshal(request)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	outboxQuery := `INSERT INTO transaction_outbox (event_type,payload, partition_key,status, correlation_id) VALUES ($1, $2, $3, $4, $5)`

	_, err = tx.Exec(ctx, outboxQuery, PaymentIntentEvent, payload, request.UserID, "pending", correlationID)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}
	return tx.Commit(ctx)

}
