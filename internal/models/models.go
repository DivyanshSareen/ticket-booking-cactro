package models

import "time"

type User struct {
	ID           int64     `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

type Event struct {
	ID               int64     `json:"id"`
	OrganizerID      int64     `json:"organizer_id"`
	Name             string    `json:"name"`
	Date             string    `json:"date"`
	Location         string    `json:"location"`
	TotalTickets     int       `json:"total_tickets"`
	AvailableTickets int       `json:"available_tickets"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type Booking struct {
	ID         int64     `json:"id"`
	CustomerID int64     `json:"customer_id"`
	EventID    int64     `json:"event_id"`
	NumTickets int       `json:"num_tickets"`
	CreatedAt  time.Time `json:"created_at"`
}
