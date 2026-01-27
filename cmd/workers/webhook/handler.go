package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Niiaks/Aegis/internal/database"
	"github.com/Niiaks/Aegis/internal/kafka"
	"github.com/Niiaks/Aegis/internal/middleware"
	"github.com/Niiaks/Aegis/internal/redis"
	"github.com/Niiaks/Aegis/pkg/constants"
	"github.com/Niiaks/Aegis/pkg/types"
	"github.com/rs/zerolog"
)

const PlatformFee int64 = 30 // 30% of the amount (store in config)
func webhookHandler(db *database.Database, redis *redis.Client, log *zerolog.Logger) kafka.Handler {
	return func(ctx context.Context, msg *kafka.Message) error {
		log.Info().Str("topic", msg.Topic).Int64("offset", msg.Offset).Msg("Processing webhook")

		var event types.PaystackWebhookEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Error().Err(err).Msg("Failed to unmarshal webhook message")
			return err
		}
		// idempotency check
		proccessed, err := redis.GetIdempotencyKey(ctx, event.Data.Reference)
		if err != nil && proccessed != "" {
			log.Info().Str("reference", event.Data.Reference).Msg("Webhook already processed,skipping")
			return nil
		}
		idStr := fmt.Sprintf("%d", event.Data.ID)
		//if processed is empty, we insert the webhook into the database
		if proccessed == "" {
			_, err := db.Pool.Exec(ctx, "INSERT INTO psp_webhooks (event_id, payload, updated_at, created_at) VALUES ($1, $2, $3, $4)", idStr, msg.Value, time.Now(), time.Now())
			if err != nil {
				log.Error().Err(err).Msg("Failed to insert webhook into database")
				return err
			}
			redis.SetIdempotencyKey(ctx, event.Data.Reference, 30*time.Minute)
		}

		// Acquire distributed lock on user wallet
		lock, err := redis.AcquireLock(ctx, "wallet:"+event.Data.Metadata.UserID, 10*time.Second)
		if err != nil {
			log.Error().Err(err).Str("user_id", event.Data.Metadata.UserID).Msg("Failed to acquire wallet lock")
			return err // Retry later
		}
		defer lock.Release(ctx)

		tx, err := db.Pool.Begin(ctx)
		if err != nil {
			log.Error().Err(err).Msg("Failed to begin transaction")
			return err
		}
		// insert into wallets(platform,seller,other), insert into ledger_entries
		// credit this, debit this so that the ledger consumer can process the ledger

		// Calculate amounts
		netAmount := event.Data.Amount - (event.Data.Amount * PlatformFee / 100)
		platformAmount := event.Data.Amount * PlatformFee / 100

		// Update seller wallet and get new balance
		var sellerBalanceAfter int64
		err = tx.QueryRow(ctx, "UPDATE wallets SET locked_balance = locked_balance + $1 WHERE user_id = $2 RETURNING locked_balance", netAmount, event.Data.Metadata.UserID).Scan(&sellerBalanceAfter)
		if err != nil {
			log.Error().Err(err).Msg("Wallet: Failed to update seller wallet")
			tx.Rollback(ctx)
			return err
		}

		// Update platform wallet and get new balance
		var platformBalanceAfter int64
		err = tx.QueryRow(ctx, "UPDATE wallets SET balance = balance + $1 WHERE id = $2 RETURNING balance", platformAmount, constants.AccountPlatformID).Scan(&platformBalanceAfter)
		if err != nil {
			log.Error().Err(err).Msg("Wallet: Failed to update platform wallet")
			tx.Rollback(ctx)
			return err
		}

		// Get external account balance (for tracking total inflows)
		var externalBalanceAfter int64
		err = tx.QueryRow(ctx, "UPDATE wallets SET balance = balance + $1 WHERE id = $2 RETURNING balance", event.Data.Amount, constants.AccountExternalID).Scan(&externalBalanceAfter)
		if err != nil {
			log.Error().Err(err).Msg("Wallet: Failed to update external wallet")
			tx.Rollback(ctx)
			return err
		}

		// Debit external account (gross amount coming in)
		_, err = tx.Exec(ctx, "INSERT INTO ledger_entries (transaction_id,account_id,debit,credit,balance_after,description,updated_at,created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", event.Data.Metadata.TransactionID, constants.AccountExternalID, event.Data.Amount, 0, externalBalanceAfter, "revenue", time.Now(), time.Now())
		if err != nil {
			log.Error().Err(err).Msg("Ledger: Failed to insert external ledger entry")
			tx.Rollback(ctx)
			return err
		}

		// Credit seller for net amount
		_, err = tx.Exec(ctx, "INSERT INTO ledger_entries (transaction_id,account_id,debit,credit,balance_after,description,updated_at,created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", event.Data.Metadata.TransactionID, event.Data.Metadata.UserID, 0, netAmount, sellerBalanceAfter, "revenue", time.Now(), time.Now())
		if err != nil {
			log.Error().Err(err).Msg("Ledger: Failed to insert seller ledger entry")
			tx.Rollback(ctx)
			return err
		}

		// Credit platform for fee
		_, err = tx.Exec(ctx, "INSERT INTO ledger_entries (transaction_id,account_id,debit,credit,balance_after,description,updated_at,created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", event.Data.Metadata.TransactionID, constants.AccountPlatformID, 0, platformAmount, platformBalanceAfter, "fee", time.Now(), time.Now())
		if err != nil {
			log.Error().Err(err).Msg("Ledger: Failed to insert ledger entry")
			tx.Rollback(ctx)
			return err
		}

		_, err = tx.Exec(ctx, `UPDATE transactions SET psp_reference = $1, status = 'completed', updated_at = NOW() WHERE id = $2`,
			event.Data.Reference, event.Data.Metadata.TransactionID)
		if err != nil {
			log.Error().Err(err).Msg("Transaction: Failed to mark as completed")
			tx.Rollback(ctx)
			return err
		}

		// Prepare balance update payload
		updateEvent := types.BalanceUpdateEvent{
			TransactionID: event.Data.Metadata.TransactionID,
			UserID:        event.Data.Metadata.UserID,
			NetAmount:     netAmount,
			Currency:      event.Data.Currency,
		}
		payloadBytes, err := json.Marshal(updateEvent)
		if err != nil {
			log.Error().Err(err).Msg("Outbox: Failed to marshal balance update event")
			tx.Rollback(ctx)
			return err
		}

		requestID := middleware.GetRequestIDFromContext(ctx)
		if requestID == "" {
			requestID = fmt.Sprintf("gen-%s", time.Now().Format("20060102150405")) // Fallback if context is lost
			log.Warn().Str("new_id", requestID).Msg("Request ID missing in context, generated fallback")
		}
		log.Info().Str("request_id", requestID).Msg("Using Correlation ID")

		_, err = tx.Exec(ctx, "INSERT INTO transaction_outbox (event_type, payload, partition_key,correlation_id, status, updated_at, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7)", kafka.EventLedgerEntryCreated, payloadBytes, event.Data.Metadata.UserID, requestID, "pending", time.Now(), time.Now())
		if err != nil {
			log.Error().Err(err).Msg("Outbox: Failed to insert ledger entry created event")
			tx.Rollback(ctx)
			return err
		}
		return tx.Commit(ctx)
	}
}
