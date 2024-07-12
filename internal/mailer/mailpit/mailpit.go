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
	GetParticipants(context.Context, uuid.UUID) ([]pgstore.Participant, error)
	GetParticipant(context.Context, uuid.UUID) (pgstore.Participant, error)
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

func (mp Mailipt) SendConfirmEmailToParticipants(tripId uuid.UUID) error {
	ctx := context.Background()

	participants, err := mp.store.GetParticipants(ctx, tripId)
	if err != nil {
		return fmt.Errorf("mailpit: failed to get trip participants for SendConfirmEmailToParticipants: %w", err)
	}

	trip, err := mp.store.GetTrip(ctx, tripId)
	if err != nil {
		return fmt.Errorf("mailpit: failed to get trip for SendConfirmEmailToParticipants: %w", err)
	}

	client, err := mail.NewClient("mailpit", mail.WithTLSPortPolicy(mail.NoTLS), mail.WithPort(1025))
	if err != nil {
		return fmt.Errorf("mailpit: failed to set client: %w", err)
	}

	for _, participant := range participants {
		msg := mail.NewMsg()
		if err := msg.From("mailpit@planner.com"); err != nil {
			return fmt.Errorf("mailpit: failed to set From in SendConfirmEmailToParticipants: %w", err)
		}

		if err := msg.To(participant.Email); err != nil {
			return fmt.Errorf("mailpit: failed to set To in SendConfirmEmailToParticipants: %w", err)
		}

		msg.Subject("Confirm your trip")
		msg.SetBodyString(mail.TypeTextPlain, fmt.Sprintf(`
			You have been invited for a trip to %s by %s.
			Click in the button below to confirm it.
		`,
			trip.Destination, trip.OwnerName,
		))

		if err := client.DialAndSend(msg); err != nil {
			return fmt.Errorf("mailpit: failed to send email: %w", err)
		}
	}

	return nil
}

func (mp Mailipt) SendConfirmEmailToInvitedParticipant(tripId uuid.UUID) error {
	ctx := context.Background()

	participant, err := mp.store.GetParticipant(ctx, tripId)
	if err != nil {
		return fmt.Errorf("mailpit: failed to get trip participants for SendConfirmEmailToInvitedParticipant: %w", err)
	}

	trip, err := mp.store.GetTrip(ctx, tripId)
	if err != nil {
		return fmt.Errorf("mailpit: failed to get trip for SendConfirmEmailToInvitedParticipant: %w", err)
	}

	msg := mail.NewMsg()
	if err := msg.From("mailpit@planner.com"); err != nil {
		return fmt.Errorf("mailpit: failed to set From in SendConfirmEmailToInvitedParticipant: %w", err)
	}
	if err := msg.To(participant.Email); err != nil {
		return fmt.Errorf("mailpit: failed to set To in SendConfirmEmailToInvitedParticipant: %w", err)
	}

	msg.Subject("Confirm your trip")
	msg.SetBodyString(mail.TypeTextPlain, fmt.Sprintf(`
		You have been invited for a trip to %s by %s.
		Click in the button below to confirm it.
	`,
		trip.Destination, trip.OwnerName,
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
