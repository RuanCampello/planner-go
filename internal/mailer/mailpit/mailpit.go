package mailpit

import (
	"context"
	"fmt"
	"planner-go/internal/pgstore"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/wneessen/go-mail"
)

type store interface {
	GetTrip(context.Context, uuid.UUID) (pgstore.Trip, error)
}

type Mailipt struct {
	store store
}

func NewMailpit(pool *pgxpool.Pool) Mailipt {
	return Mailipt{pgstore.New(pool)}
}

func (mp Mailipt) SendConfirmEmailToTripOwner(tripId uuid.UUID) error {
	ctx := context.Background()

	trip, err := mp.store.GetTrip(ctx, tripId)
	if err != nil {
		return fmt.Errorf("mailpit: failed to get trip for SendConfirmEmailToTripOwner: %w", err)
	}

	msg := mail.NewMsg()
	if err := msg.From("mailpit@planner.com"); err != nil {
		return fmt.Errorf("mailpit: failed to set From in SendConfirmEmailToTripOwner: %w", err)
	}

	if err := msg.To(trip.OwnerEmail); err != nil {
		return fmt.Errorf("mailpit: failed to set To in SendConfirmEmailToTripOwner: %w", err)
	}

	msg.Subject("Confirm your trip")
	msg.SetBodyString(mail.TypeTextPlain, fmt.Sprintf(`
		Hello, %s!
		Your trip to %s which starts on %s needs to be confirmed.
		Click in the button below to confirm it.
	`,
		trip.OwnerName, trip.Destination, trip.StartsAt.Time.Format("02-01-2006"),
	))

	client, err := mail.NewClient("mailpit", mail.WithTLSPortPolicy(mail.NoTLS), mail.WithPort(1025))

	if err != nil {
		return fmt.Errorf("mailpit: failed to set client: %w", err)
	}

	if err := client.DialAndSend(msg); err != nil {
		return fmt.Errorf("mailpit: failed to send email: %w", err)
	}

	return nil
}
