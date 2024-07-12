package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"planner-go/internal/api/spec"
	"planner-go/internal/pgstore"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type store interface {
	GetParticipant(context.Context, uuid.UUID) (pgstore.Participant, error)
	ConfirmParticipant(context.Context, uuid.UUID) error
	CreateTrip(context.Context, *pgxpool.Pool, spec.CreateTripRequest) (uuid.UUID, error)
}

type mailer interface {
	SendConfirmEmailToTripOwner(uuid.UUID) error
}

type API struct {
	store     store
	logger    *zap.Logger
	validator *validator.Validate
	pool      *pgxpool.Pool
	mailer    mailer
}

func NewApi(pool *pgxpool.Pool, logger *zap.Logger, mailer mailer) API {
	validator := validator.New(validator.WithRequiredStructEnabled())
	return API{pgstore.New(pool), logger, validator, pool, mailer}
}

// Confirms a participant on a trip.
// (PATCH /participants/{participantId}/confirm)
func (api API) PatchParticipantsParticipantIDConfirm(w http.ResponseWriter, r *http.Request, participantID string) *spec.Response {
	id, err := uuid.Parse(participantID)
	if err != nil {
		return spec.PatchParticipantsParticipantIDConfirmJSON400Response(spec.Error{Message: "Invalid UUID"})
	}

	participant, err := api.store.GetParticipant(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return spec.PatchParticipantsParticipantIDConfirmJSON400Response(spec.Error{Message: "Participant not found"})
		}
		api.logger.Error("Failed to get participant", zap.Error(err), zap.String("participant_id", participantID))
		return spec.PatchParticipantsParticipantIDConfirmJSON400Response(spec.Error{Message: "Something went wrong"})
	}

	if participant.IsConfirmed {
		return spec.PatchParticipantsParticipantIDConfirmJSON400Response(spec.Error{Message: "Participant already confirmed"})
	}

	if err := api.store.ConfirmParticipant(r.Context(), id); err != nil {
		api.logger.Error("Failed to confirm participant", zap.Error(err), zap.String("participant_id", participantID))
		return spec.PatchParticipantsParticipantIDConfirmJSON400Response(spec.Error{Message: "Something went wrong"})
	}

	return spec.PatchParticipantsParticipantIDConfirmJSON204Response(nil)
}

// Create a new trip
// (POST /trips)
func (api API) PostTrips(w http.ResponseWriter, r *http.Request) *spec.Response {
	var body spec.CreateTripRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return spec.PostTripsJSON400Response(spec.Error{Message: "Invalid JSON"})
	}

	if err := api.validator.Struct(body); err != nil {
		return spec.PostTripsJSON400Response(spec.Error{Message: "Invalid input" + err.Error()})
	}

	tripId, err := api.store.CreateTrip(r.Context(), api.pool, body)

	if err != nil {
		return spec.PostTripsJSON400Response(spec.Error{Message: "Something went wrong"})
	}

	go func() {
		if err := api.mailer.SendConfirmEmailToTripOwner(tripId); err != nil {
			api.logger.Error("Failed to send email on PostTrips",
				zap.Error(err),
				zap.String("trip_id", tripId.String()))
		}
	}()

	return spec.PostTripsJSON201Response(spec.CreateTripResponse{TripID: tripId.String()})
}

// Get a trip details.
// (GET /trips/{tripId})
func (api API) GetTripsTripID(w http.ResponseWriter, r *http.Request, tripID string) *spec.Response {
	panic("not implemented") // TODO: Implement
}

// Update a trip.
// (PUT /trips/{tripId})
func (api API) PutTripsTripID(w http.ResponseWriter, r *http.Request, tripID string) *spec.Response {
	panic("not implemented") // TODO: Implement
}

// Get a trip activities.
// (GET /trips/{tripId}/activities)
func (api API) GetTripsTripIDActivities(w http.ResponseWriter, r *http.Request, tripID string) *spec.Response {
	panic("not implemented") // TODO: Implement
}

// Create a trip activity.
// (POST /trips/{tripId}/activities)
func (api API) PostTripsTripIDActivities(w http.ResponseWriter, r *http.Request, tripID string) *spec.Response {
	panic("not implemented") // TODO: Implement
}

// Confirm a trip and send e-mail invitations.
// (GET /trips/{tripId}/confirm)
func (api API) GetTripsTripIDConfirm(w http.ResponseWriter, r *http.Request, tripID string) *spec.Response {
	panic("not implemented") // TODO: Implement
}

// Invite someone to the trip.
// (POST /trips/{tripId}/invites)
func (api API) PostTripsTripIDInvites(w http.ResponseWriter, r *http.Request, tripID string) *spec.Response {
	panic("not implemented") // TODO: Implement
}

// Get a trip links.
// (GET /trips/{tripId}/links)
func (api API) GetTripsTripIDLinks(w http.ResponseWriter, r *http.Request, tripID string) *spec.Response {
	panic("not implemented") // TODO: Implement
}

// Create a trip link.
// (POST /trips/{tripId}/links)
func (api API) PostTripsTripIDLinks(w http.ResponseWriter, r *http.Request, tripID string) *spec.Response {
	panic("not implemented") // TODO: Implement
}

// Get a trip participants.
// (GET /trips/{tripId}/participants)
func (api API) GetTripsTripIDParticipants(w http.ResponseWriter, r *http.Request, tripID string) *spec.Response {
	panic("not implemented") // TODO: Implement
}
