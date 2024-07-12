package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/mail"
	"planner-go/internal/api/spec"
	"planner-go/internal/pgstore"
	"strings"
	"time"

	"github.com/discord-gophers/goapi-gen/types"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type store interface {
	//trip functions
	GetTrip(context.Context, uuid.UUID) (pgstore.Trip, error)
	CreateTrip(context.Context, *pgxpool.Pool, spec.CreateTripRequest) (uuid.UUID, error)
	UpdateTrip(context.Context, pgstore.UpdateTripParams) error
	InviteParticipantToTrip(context.Context, pgstore.InviteParticipantToTripParams) (uuid.UUID, error)
	//participant functions
	GetParticipant(context.Context, uuid.UUID) (pgstore.Participant, error)
	GetParticipants(context.Context, uuid.UUID) ([]pgstore.Participant, error)
	ConfirmParticipant(context.Context, uuid.UUID) error
	//activities functions
	GetTripActivities(context.Context, uuid.UUID) ([]pgstore.Activity, error)
	CreateActivity(context.Context, pgstore.CreateActivityParams) (uuid.UUID, error)
}

type mailer interface {
	SendConfirmEmailToTripOwner(uuid.UUID) error
	SendConfirmEmailToParticipants(uuid.UUID) error
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
		return spec.PostTripsJSON400Response(spec.Error{Message: "Invalid JSON Body"})
	}

	if err := api.validator.Struct(body); err != nil {
		return spec.PostTripsJSON400Response(spec.Error{Message: "Invalid input field" + err.Error()})
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
	id, err := uuid.Parse(tripID)
	if err != nil {
		return spec.GetTripsTripIDJSON400Response(spec.Error{Message: "Invalid UUId"})
	}

	trip, err := api.store.GetTrip(r.Context(), id)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return spec.GetTripsTripIDJSON400Response(spec.Error{Message: "Trip not found"})
		}
		api.logger.Error("Failed to get trip", zap.Error(err), zap.String("trip_id", tripID))
		return spec.PatchParticipantsParticipantIDConfirmJSON400Response(spec.Error{Message: "Something went wrong"})
	}

	return spec.GetTripsTripIDJSON200Response(spec.GetTripDetailsResponse{Trip: spec.GetTripDetailsResponseTripObj{
		ID:          tripID,
		Destination: trip.Destination,
		StartsAt:    trip.StartsAt.Time,
		EndsAt:      trip.EndsAt.Time,
		IsConfirmed: trip.IsConfirmed,
	}})
}

// Update a trip.
// (PUT /trips/{tripId})
func (api API) PutTripsTripID(w http.ResponseWriter, r *http.Request, tripID string) *spec.Response {
	var body spec.PutTripsTripIDJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return spec.PostTripsJSON400Response(spec.Error{Message: "Invalid JSON Body"})
	}

	if err := api.validator.Struct(body); err != nil {
		return spec.PutTripsTripIDJSON400Response(spec.Error{Message: "Invalid input field" + err.Error()})
	}

	id, err := uuid.Parse(tripID)
	if err != nil {
		return spec.PutTripsTripIDJSON400Response(spec.Error{Message: "Invalid UUID"})
	}

	trip, err := api.store.GetTrip(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return spec.GetTripsTripIDJSON400Response(spec.Error{Message: "Trip not found"})
		}
		api.logger.Error("Failed to get trip", zap.Error(err), zap.String("trip_id", tripID))
		return spec.PutTripsTripIDJSON400Response(spec.Error{Message: "Something went wrong"})
	}

	if err := api.store.UpdateTrip(r.Context(), pgstore.UpdateTripParams{
		Destination: body.Destination,
		StartsAt:    pgtype.Timestamp{Time: body.StartsAt, Valid: true},
		EndsAt:      pgtype.Timestamp{Time: body.EndsAt, Valid: true},
		IsConfirmed: trip.IsConfirmed,
		ID:          id,
	}); err != nil {
		api.logger.Error("Failed to update trip", zap.Error(err), zap.String("trip_id", tripID))
		return spec.PutTripsTripIDJSON400Response(spec.Error{Message: "Something went wrong"})
	}

	return spec.PutTripsTripIDJSON204Response(nil)
}

// Get a trip activities.
// (GET /trips/{tripId}/activities)
func (api API) GetTripsTripIDActivities(w http.ResponseWriter, r *http.Request, tripID string) *spec.Response {
	id, err := uuid.Parse(tripID)
	if err != nil {
		return spec.GetTripsTripIDActivitiesJSON400Response(spec.Error{Message: "Invalid UUID"})
	}

	activities, err := api.store.GetTripActivities(r.Context(), id)

	if !(len(activities) > 0) {
		return spec.GetTripsTripIDActivitiesJSON400Response(spec.Error{Message: "No activities found"})
	}

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return spec.GetTripsTripIDActivitiesJSON400Response(spec.Error{Message: "Trip not found"})
		}
		api.logger.Error("Failed to get trip participants", zap.Error(err), zap.String("trip_id", tripID))
		return spec.GetTripsTripIDActivitiesJSON400Response(spec.Error{Message: "Something went wrong"})
	}

	var response spec.GetTripActivitiesResponse

	//group activities by date
	groupedAct := make(map[string][]pgstore.Activity)

	for _, activity := range activities {
		date := activity.OccursAt.Time.Format(time.DateOnly)
		groupedAct[date] = append(groupedAct[date], activity)
	}

	//format activities for response
	for dtString, actsOnDate := range groupedAct {
		var inActs []spec.GetTripActivitiesResponseInnerArray

		for _, act := range actsOnDate {
			inActs = append(inActs, spec.GetTripActivitiesResponseInnerArray{
				ID:       act.ID.String(),
				Title:    act.Title,
				OccursAt: act.OccursAt.Time,
			})
		}

		date, _ := time.Parse(time.DateOnly, dtString)
		response.Activities = append(response.Activities, spec.GetTripActivitiesResponseOuterArray{
			Date:       date,
			Activities: inActs,
		})
	}

	return spec.GetTripsTripIDActivitiesJSON200Response(response)
}

// Create a trip activity.
// (POST /trips/{tripId}/activities)
func (api API) PostTripsTripIDActivities(w http.ResponseWriter, r *http.Request, tripID string) *spec.Response {
	var body spec.PostTripsTripIDActivitiesJSONRequestBody

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return spec.GetTripsTripIDActivitiesJSON400Response(spec.Error{Message: "Invalid JSON Body"})
	}

	if err := api.validator.Struct(body); err != nil {
		return spec.GetTripsTripIDActivitiesJSON400Response(spec.Error{Message: "Invalid input field" + err.Error()})
	}

	id, err := uuid.Parse(tripID)
	if err != nil {
		return spec.PostTripsTripIDActivitiesJSON400Response(spec.Error{Message: "Invalid UUID"})
	}

	activityId, err := api.store.CreateActivity(r.Context(), pgstore.CreateActivityParams{
		TripID:   id,
		Title:    body.Title,
		OccursAt: pgtype.Timestamp{Time: body.OccursAt, Valid: true},
	})

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			spec.PostTripsTripIDActivitiesJSON400Response(spec.Error{Message: "Trip not found"})
		}
		api.logger.Error("Failed to create activity", zap.Error(err), zap.String("trip_id", tripID))
		return spec.PostTripsTripIDActivitiesJSON400Response(spec.Error{Message: "Something went wrong"})
	}

	return spec.PostTripsTripIDActivitiesJSON201Response(spec.CreateActivityResponse{ActivityID: activityId.String()})
}

// Confirm a trip and send e-mail invitations.
// (GET /trips/{tripId}/confirm)
func (api API) GetTripsTripIDConfirm(w http.ResponseWriter, r *http.Request, tripID string) *spec.Response {
	id, err := uuid.Parse(tripID)
	if err != nil {
		spec.GetTripsTripIDConfirmJSON400Response(spec.Error{Message: "Invalid UUID"})
	}

	trip, err := api.store.GetTrip(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			spec.GetTripsTripIDConfirmJSON400Response(spec.Error{Message: "Trip not found"})
		}
		api.logger.Error("Failed to get trip", zap.Error(err), zap.String("trip_id", tripID))
		return spec.GetTripsTripIDConfirmJSON400Response(spec.Error{Message: "Something went wrong"})
	}

	if trip.IsConfirmed {
		return spec.GetTripsTripIDConfirmJSON400Response(spec.Error{Message: "Trip is already confirmed"})
	}

	if err := api.store.UpdateTrip(r.Context(), pgstore.UpdateTripParams{
		ID:          id,
		Destination: trip.Destination,
		EndsAt:      trip.EndsAt,
		StartsAt:    trip.StartsAt,
		IsConfirmed: true,
	}); err != nil {
		api.logger.Error("Failed to update trip", zap.Error(err), zap.String("trip_id", tripID))
		return spec.GetTripsTripIDConfirmJSON400Response(spec.Error{Message: "Something went wrong"})
	}

	go func() {
		if err := api.mailer.SendConfirmEmailToParticipants(id); err != nil {
			api.logger.Error("Failed to send email on GetTripsTripIDConfirm",
				zap.Error(err),
				zap.String("trip_id", tripID))
		}
	}()

	return spec.GetTripsTripIDConfirmJSON204Response(nil)
}

// Invite someone to the trip.
// (POST /trips/{tripId}/invites)
func (api API) PostTripsTripIDInvites(w http.ResponseWriter, r *http.Request, tripID string) *spec.Response {
	var body spec.PostTripsTripIDInvitesJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return spec.PostTripsTripIDInvitesJSON400Response(spec.Error{Message: "Invalid JSON Body"})
	}

	id, err := uuid.Parse(tripID)
	if err != nil {
		return spec.PostTripsTripIDInvitesJSON400Response(spec.Error{Message: "Invalid UUID"})
	}

	participantId, err := api.store.InviteParticipantToTrip(r.Context(), pgstore.InviteParticipantToTripParams{
		TripID: id,
		Email:  string(body.Email),
	})

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			spec.PostTripsTripIDInvitesJSON400Response(spec.Error{Message: "Trip not found"})
		}
		api.logger.Error("Failed to invite participant for the trip",
			zap.Error(err),
			zap.String("trip_id", tripID),
			zap.String("participant_id", participantId.String()),
			zap.String("participant_email", string(body.Email)),
		)
		return spec.PostTripsTripIDInvitesJSON400Response(spec.Error{Message: "Something went wrong"})
	}

	//TODO: send email to invited participant

	return spec.PostTripsTripIDInvitesJSON201Response(nil)
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
	id, err := uuid.Parse(tripID)
	if err != nil {
		return spec.GetTripsTripIDParticipantsJSON400Response(spec.Error{Message: "Invalid UUID"})
	}

	participants, err := api.store.GetParticipants(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return spec.GetTripsTripIDParticipantsJSON400Response(spec.Error{Message: "No trip found"})
		}
		api.logger.Error("Failed to get trip participants", zap.Error(err), zap.String("trip_id", tripID))
		return spec.GetTripsTripIDParticipantsJSON400Response(spec.Error{Message: "Something went wrong"})
	}

	var response spec.GetTripParticipantsResponse
	response.Participants = make([]spec.GetTripParticipantsResponseArray, len(participants))

	for i, participant := range participants {
		var name string
		formattedEmail, err := mail.ParseAddress(participant.Email)
		if err == nil {
			addr := formattedEmail.Address
			name = addr[:strings.Index(addr, "@")]
		}
		response.Participants[i] = spec.GetTripParticipantsResponseArray{
			ID:          participant.ID.String(),
			Email:       types.Email(participant.Email),
			IsConfirmed: participant.IsConfirmed,
			Name:        &name,
		}
	}

	return spec.GetTripsTripIDParticipantsJSON200Response(response)
}
