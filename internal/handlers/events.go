package handlers

import (
	"database/sql"
	"net/http"

	"ticket-booking/internal/db"
	"ticket-booking/internal/models"
	"ticket-booking/internal/worker"

	"github.com/gin-gonic/gin"
)

type createEventRequest struct {
	Name         string `json:"name"          binding:"required"`
	Date         string `json:"date"          binding:"required"`
	Location     string `json:"location"      binding:"required"`
	TotalTickets int    `json:"total_tickets"`
}

func CreateEvent(c *gin.Context) {
	var req createEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	organizerID := c.GetInt64("user_id")
	totalTickets := req.TotalTickets
	if totalTickets <= 0 {
		totalTickets = 50
	}

	res, err := db.DB.Exec(
		`INSERT INTO events (organizer_id, name, date, location, total_tickets, available_tickets)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		organizerID, req.Name, req.Date, req.Location, totalTickets, totalTickets,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create event"})
		return
	}

	id, _ := res.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{
		"id": id, "name": req.Name, "date": req.Date,
		"location": req.Location, "total_tickets": totalTickets, "available_tickets": totalTickets,
	})
}

type updateEventRequest struct {
	Name     *string `json:"name"`
	Date     *string `json:"date"`
	Location *string `json:"location"`
}

func UpdateEvent(c *gin.Context) {
	eventID := c.Param("id")
	organizerID := c.GetInt64("user_id")

	var event models.Event
	err := db.DB.QueryRow(
		`SELECT id, organizer_id, name, date, location, total_tickets, available_tickets, created_at, updated_at FROM events WHERE id = ?`, eventID,
	).Scan(&event.ID, &event.OrganizerID, &event.Name, &event.Date, &event.Location, &event.TotalTickets, &event.AvailableTickets, &event.CreatedAt, &event.UpdatedAt)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
		return
	}
	if event.OrganizerID != organizerID {
		c.JSON(http.StatusForbidden, gin.H{"error": "you don't own this event"})
		return
	}

	var req updateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Name != nil {
		event.Name = *req.Name
	}
	if req.Date != nil {
		event.Date = *req.Date
	}
	if req.Location != nil {
		event.Location = *req.Location
	}

	_, err = db.DB.Exec(
		`UPDATE events SET name=?, date=?, location=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`,
		event.Name, event.Date, event.Location, event.ID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not update event"})
		return
	}
	db.DB.QueryRow(`SELECT updated_at FROM events WHERE id=?`, event.ID).Scan(&event.UpdatedAt)

	// collect booked customers and fire notification job
	rows, _ := db.DB.Query(
		`SELECT DISTINCT u.email FROM bookings b
		 JOIN users u ON u.id = b.customer_id
		 WHERE b.event_id = ?`, event.ID,
	)
	var customers []string
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var email string
			rows.Scan(&email)
			customers = append(customers, email)
		}
	}

	if len(customers) > 0 {
		worker.Enqueue(worker.Job{
			Type: worker.JobEventUpdateNotify,
			Payload: map[string]any{
				"event_name": event.Name,
				"customers":  customers,
			},
		})
	}

	c.JSON(http.StatusOK, event)
}

func ListEvents(c *gin.Context) {
	rows, err := db.DB.Query(
		`SELECT id, organizer_id, name, date, location, total_tickets, available_tickets, created_at, updated_at FROM events ORDER BY date`,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch events"})
		return
	}
	defer rows.Close()

	var events []models.Event
	for rows.Next() {
		var e models.Event
		rows.Scan(&e.ID, &e.OrganizerID, &e.Name, &e.Date, &e.Location, &e.TotalTickets, &e.AvailableTickets, &e.CreatedAt, &e.UpdatedAt)
		events = append(events, e)
	}
	if events == nil {
		events = []models.Event{}
	}
	c.JSON(http.StatusOK, events)
}

func GetEvent(c *gin.Context) {
	eventID := c.Param("id")
	var e models.Event
	err := db.DB.QueryRow(
		`SELECT id, organizer_id, name, date, location, total_tickets, available_tickets, created_at, updated_at FROM events WHERE id = ?`, eventID,
	).Scan(&e.ID, &e.OrganizerID, &e.Name, &e.Date, &e.Location, &e.TotalTickets, &e.AvailableTickets, &e.CreatedAt, &e.UpdatedAt)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
		return
	}
	c.JSON(http.StatusOK, e)
}
