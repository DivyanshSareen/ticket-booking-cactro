# Ticket Booking API

A REST API for an event booking system built with Go, Gin, SQLite, and JWT auth.

## Features

- Two user roles: **Organizer** and **Customer**
- Organizers create and update events
- Customers browse events and book tickets
- Atomic ticket booking — no overselling
- Background worker pool for async jobs (booking confirmation, event update notifications)

## Stack

- **Go** + **Gin** — HTTP framework
- **SQLite** — embedded database
- **JWT** — stateless auth
- **bcrypt** — password hashing
- Channel-based worker pool — background tasks

## Setup

**Prerequisites:** Go 1.21+, gcc (required for SQLite CGo bindings)

```bash
git clone <repo-url>
cd ticket-booking-cactro

cp .env.example .env
# Edit .env and set a strong JWT_SECRET

go mod download
go run .
```

The server starts on `http://localhost:8080` (configurable via `PORT` in `.env`).

## Environment Variables

| Variable     | Description              | Default                        |
|--------------|--------------------------|--------------------------------|
| `JWT_SECRET` | Secret key for JWT signing | `change-me-in-production`    |
| `PORT`       | Port to listen on        | `8080`                         |

## Postman Collection

[![Run in Postman](https://run.pstmn.io/button.svg)](https://www.postman.com/martian-trinity-154020/kiroween/collection/doh9gg6/ticket-booking-api?action=share&source=copy-link&creator=0)

---

## API Reference

### Auth

| Method | Endpoint          | Auth     | Description        |
|--------|-------------------|----------|--------------------|
| POST   | `/auth/register`  | None     | Register a user    |
| POST   | `/auth/login`     | None     | Login, get token   |

**Register body:**
```json
{ "email": "user@example.com", "password": "secret123", "role": "organizer" }
```
`role` must be `organizer` or `customer`.

**Login response:**
```json
{ "token": "<jwt>", "role": "organizer" }
```

Use the token as `Authorization: Bearer <token>` on all subsequent requests.

---

### Events

| Method | Endpoint        | Role       | Description          |
|--------|-----------------|------------|----------------------|
| GET    | `/events`       | Any        | List all events      |
| GET    | `/events/:id`   | Any        | Get a single event   |
| POST   | `/events`       | Organizer  | Create an event      |
| PUT    | `/events/:id`   | Organizer  | Update own event     |

**Create event body:**
```json
{ "name": "Rock Concert", "date": "2026-06-01", "location": "Mumbai Arena", "total_tickets": 100 }
```
`total_tickets` defaults to 50 if omitted.

---

### Bookings

| Method | Endpoint              | Role     | Description              |
|--------|-----------------------|----------|--------------------------|
| POST   | `/events/:id/book`    | Customer | Book tickets for event   |
| GET    | `/bookings`           | Customer | List your bookings       |

**Book body:**
```json
{ "num_tickets": 2 }
```

Booking is all-or-nothing — if fewer tickets remain than requested, the request is rejected with `409 Conflict`.

---

## Background Jobs

Two async jobs run via a 3-worker pool:

- **Booking confirmation** — logged to console when a customer books tickets
- **Event update notification** — logged to console when an organizer updates an event, listing all affected customers
