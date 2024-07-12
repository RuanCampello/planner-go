package pgstore

import (
	"context"
	"fmt"
	"planner-go/internal/api/spec"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

func (q *Queries) CreateTrip(ctx context.Context, pool *pgxpool.Pool, params spec.CreateTripRequest) (uuid.UUID, error) {
	tx, err := pool.Begin(ctx)

	if err != nil {
		return uuid.UUID{}, fmt.Errorf("pgstore: failed to begin trx for CreatTrip: %w", err)
	}

	defer tx.Rollback(ctx)

	qtx := q.WithTx(tx)

	tripId, err := qtx.InsertTrip(ctx, InsertTripParams{
		Destination: params.Destination,
		OwnerEmail:  string(params.OwnerEmail),
		OwnerName:   params.OwnerName,
		StartsAt:    pgtype.Timestamp{Valid: true, Time: params.StartsAt},
		EndsAt:      pgtype.Timestamp{Valid: true, Time: params.EndsAt},
	})

	if err != nil {
		return uuid.UUID{}, fmt.Errorf("pgstore: failed to insert trip for CreatTrip: %w", err)
	}

	participants := make([]InviteParticipantsToTripParams, len(params.EmailsToInvite))
	for i, eti := range params.EmailsToInvite {
		participants[i] = InviteParticipantsToTripParams{
			TripID: tripId,
			Email:  string(eti),
		}
	}

	if _, err := qtx.InviteParticipantsToTrip(ctx, participants); err != nil {
		return uuid.UUID{}, fmt.Errorf("pgstore: failed to invite participants for CreatTrip: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return uuid.UUID{}, fmt.Errorf("pgstore: failed to begin trx for CreatTrip: %w", err)
	}

	return tripId, nil
}
