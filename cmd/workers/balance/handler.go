package main

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Niiaks/Aegis/internal/database"
	"github.com/Niiaks/Aegis/internal/kafka"
	"github.com/Niiaks/Aegis/internal/redis"
	"github.com/Niiaks/Aegis/pkg/types"
	"github.com/rs/zerolog"
)

func balanceHandler(db *database.Database, redis *redis.Client, log *zerolog.Logger) kafka.Handler {
	return func(ctx context.Context, msg *kafka.Message) error {
		log.Info().Str("topic", msg.Topic).Int64("offset", msg.Offset).Msg("Processing balance update")

		var event types.BalanceUpdateEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Error().Err(err).Msg("Failed to unmarshal balance update message")
			return err
		}

		// Validation: skip old messages with full webhook JSON or empty values
		// Remove later
		if event.UserID == "" || event.NetAmount <= 0 {
			log.Warn().
				Int64("offset", msg.Offset).
				Str("user_id", event.UserID).
				Int64("amount", event.NetAmount).
				Msg("Skipping invalid or old balance update payload")
			return nil
		}

		// Acquire lock on user wallet
		lock, err := redis.AcquireLock(ctx, "wallet:"+event.UserID, 10*time.Second)
		if err != nil {
			log.Error().Err(err).Str("user_id", event.UserID).Msg("Failed to acquire wallet lock")
			return err
		}
		defer lock.Release(ctx)

		// Atomically move funds from locked_balance to balance
		// We use a check on locked_balance >= amount to ensure we don't go negative (idempotency/safety check)
		res, err := db.Pool.Exec(ctx, `
			UPDATE wallets 
			SET locked_balance = locked_balance - $1, 
				balance = balance + $1, 
				updated_at = NOW() 
			WHERE user_id = $2 AND locked_balance >= $1`,
			event.NetAmount, event.UserID)

		if err != nil {
			log.Error().Err(err).Str("user_id", event.UserID).Msg("Failed to finalize balance move")
			return err
		}

		if res.RowsAffected() == 0 {
			log.Warn().Str("user_id", event.UserID).Int64("amount", event.NetAmount).Msg("No rows updated. Balance may have already been moved or insufficient locked funds.")
		} else {
			log.Info().Str("user_id", event.UserID).Int64("amount", event.NetAmount).Msg("Successfully finalized balance move")
		}

		return nil
	}
}
