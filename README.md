# plann.er API Documentation

Welcome to the documentation for the plann.er API. This API is designed to manage trip planning, including creating trips, inviting participants, and managing activities and links associated with trips.

## Table of Contents
- [Overview](#overview)
- [Running the API](#running-the-api)
- [Endpoints](#endpoints)
  - [Confirm Trip](#confirm-trip)
  - [Confirm Participant](#confirm-participant)
  - [Invite Participant](#invite-participant)
  - [Create Trip Activity](#create-trip-activity)
  - [Get Trip Activities](#get-trip-activities)
  - [Create Trip Link](#create-trip-link)
  - [Get Trip Links](#get-trip-links)
  - [Create Trip](#create-trip)
  - [Get Trip Details](#get-trip-details)
  - [Update Trip](#update-trip)
  - [Get Trip Participants](#get-trip-participants)

## Overview
The plann.er API allows you to manage trips, invite participants, and handle various activities and links related to trips. Each endpoint is documented with example requests and responses to guide you in using the API effectively.

## Running the API
To run the plann.er API, you can use Docker Compose. Follow these steps to get the API up and running:

1. **Clone the Repository**
   
   Clone the repository containing the Docker Compose configuration.

   ```sh
   git clone https://github.com/RuanCampello/planner-go.git
   ```

2. **Create a `.env` File**
   
   Create a `.env` file in the root directory of the repository with the necessary environment variables. Here is an example:

   ```env
   PLANNER_DATABASE_HOST=localhost
   PLANNER_DATABASE_PORT=5432
   PLANNER_DATABASE_NAME=
   PLANNER_DATABASE_USER=
   PLANNER_DATABASE_PASSWORD=
   ```

3. **Run Docker Compose**
   
   Use Docker Compose to build and run the containers.

   ```sh
   docker-compose up --build
   ```

   This command will start the API and its dependencies (e.g., database) in Docker containers.

4. **Access the API**

   Once the containers are running, you can access the API at `http://localhost:8000`.

## Endpoints

### Confirm Trip
**Endpoint:** `GET /trips/{tripId}/confirm`

**Description:** Confirm a trip and send e-mail invitations.

**Path Parameters:**
- `tripId` (string, uuid): The ID of the trip to confirm.

**Responses:**

- **204 No Content**

  Example Response:
  ```json
  null
  ```

- **400 Bad Request**

  Example Response:
  ```json
  {
    "message": "Invalid trip ID."
  }
  ```

---

### Confirm Participant
**Endpoint:** `PATCH /participants/{participantId}/confirm`

**Description:** Confirms a participant on a trip.

**Path Parameters:**
- `participantId` (string, uuid): The ID of the participant to confirm.

**Responses:**

- **204 No Content**

  Example Response:
  ```json
  null
  ```

- **400 Bad Request**

  Example Response:
  ```json
  {
    "message": "Invalid participant ID."
  }
  ```

---

### Invite Participant
**Endpoint:** `POST /trips/{tripId}/invites`

**Description:** Invite someone to the trip.

**Path Parameters:**
- `tripId` (string, uuid): The ID of the trip to which the participant is invited.

**Request Body:**
```json
{
  "email": "invitee@example.com"
}
```

**Responses:**

- **201 Created**

  Example Response:
  ```json
  null
  ```

- **400 Bad Request**

  Example Response:
  ```json
  {
    "message": "Invalid email format."
  }
  ```

---

### Create Trip Activity
**Endpoint:** `POST /trips/{tripId}/activities`

**Description:** Create a trip activity.

**Path Parameters:**
- `tripId` (string, uuid): The ID of the trip for which the activity is created.

**Request Body:**
```json
{
  "occurs_at": "2024-07-15T10:00:00Z",
  "title": "City Tour"
}
```

**Responses:**

- **201 Created**

  Example Response:
  ```json
  {
    "activityId": "123e4567-e89b-12d3-a456-426614174000"
  }
  ```

- **400 Bad Request**

  Example Response:
  ```json
  {
    "message": "Invalid date format."
  }
  ```

---

### Get Trip Activities
**Endpoint:** `GET /trips/{tripId}/activities`

**Description:** Get a trip activities.

**Path Parameters:**
- `tripId` (string, uuid): The ID of the trip.

**Responses:**

- **200 OK**

  Example Response:
  ```json
  {
    "activities": [
      {
        "date": "2024-07-15T00:00:00Z",
        "activities": [
          {
            "id": "123e4567-e89b-12d3-a456-426614174001",
            "title": "City Tour",
            "occurs_at": "2024-07-15T10:00:00Z"
          }
        ]
      }
    ]
  }
  ```

- **400 Bad Request**

  Example Response:
  ```json
  {
    "message": "Invalid trip ID."
  }
  ```

---

### Create Trip Link
**Endpoint:** `POST /trips/{tripId}/links`

**Description:** Create a trip link.

**Path Parameters:**
- `tripId` (string, uuid): The ID of the trip for which the link is created.

**Request Body:**
```json
{
  "title": "Booking Link",
  "url": "https://booking.com/trip/123"
}
```

**Responses:**

- **201 Created**

  Example Response:
  ```json
  {
    "linkId": "123e4567-e89b-12d3-a456-426614174002"
  }
  ```

- **400 Bad Request**

  Example Response:
  ```json
  {
    "message": "Invalid URL format."
  }
  ```

---

### Get Trip Links
**Endpoint:** `GET /trips/{tripId}/links`

**Description:** Get a trip links.

**Path Parameters:**
- `tripId` (string, uuid): The ID of the trip.

**Responses:**

- **200 OK**

  Example Response:
  ```json
  {
    "links": [
      {
        "id": "123e4567-e89b-12d3-a456-426614174002",
        "title": "Booking Link",
        "url": "https://booking.com/trip/123"
      }
    ]
  }
  ```

- **400 Bad Request**

  Example Response:
  ```json
  {
    "message": "Invalid trip ID."
  }
  ```

---

### Create Trip
**Endpoint:** `POST /trips`

**Description:** Create a new trip.

**Request Body:**
```json
{
  "destination": "New York",
  "starts_at": "2024-07-20T00:00:00Z",
  "ends_at": "2024-07-25T00:00:00Z",
  "emails_to_invite": ["invitee1@example.com", "invitee2@example.com"],
  "owner_name": "John Doe",
  "owner_email": "john.doe@example.com"
}
```

**Responses:**

- **201 Created**

  Example Response:
  ```json
  {
    "tripId": "123e4567-e89b-12d3-a456-426614174003"
  }
  ```

- **400 Bad Request**

  Example Response:
  ```json
  {
    "message": "Invalid input data."
  }
  ```

---

### Get Trip Details
**Endpoint:** `GET /trips/{tripId}`

**Description:** Get a trip details.

**Path Parameters:**
- `tripId` (string, uuid): The ID of the trip.

**Responses:**

- **200 OK**

  Example Response:
  ```json
  {
    "trip": {
      "id": "123e4567-e89b-12d3-a456-426614174003",
      "destination": "New York",
      "starts_at": "2024-07-20T00:00:00Z",
      "ends_at": "2024-07-25T00:00:00Z",
      "is_confirmed": true
    }
  }
  ```

- **400 Bad Request**

  Example Response:
  ```json
  {
    "message": "Invalid trip ID."
  }
  ```

---

### Update Trip
**Endpoint:** `PUT /trips/{tripId}`

**Description:** Update a trip.

**Path Parameters:**
- `tripId` (string, uuid): The ID of the trip to update.

**Request Body:**
```json
{
  "destination": "Los Angeles",
  "starts_at": "2024-08-01T00:00:00Z",
  "ends_at": "2024-08-05T00:00:00Z"
}
```

**Responses:**

- **200 OK**

  Example Response:
  ```json
  {
    "tripId": "123e4567-e89b-12d3-a456-426614174003"
  }
  ```

- **400 Bad Request**

  Example Response:
  ```json
  {
    "message": "Invalid input data."
  }
  ```

---

### Get Trip Participants
**Endpoint:** `GET /trips/{tripId}/participants`

**Description:** Get the participants of a trip.

**Path Parameters:**
- `tripId` (string, uuid): The ID of the trip.

**Responses:**

- **200 OK**

  Example Response:
  ```json
  {
    "participants": [
      {
        "id": "123e4567-e89b-12d3-a456-426614174004",
        "email": "invitee1@example.com",
        "name": "Alice",
        "confirmed": true
      }
    ]
  }
  ```

- **400 Bad Request**

  Example Response:
  ```json
  {
    "message": "Invalid trip ID."
  }
  ```