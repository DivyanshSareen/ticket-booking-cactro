package handlers

import (
	"database/sql"
	"net/http"

	"ticket-booking/internal/db"
	"ticket-booking/internal/models"
	"ticket-booking/internal/worker"

	"github.com/gin-gonic/gin"
)

type bookRequest struct {
	NumTickets int `json:"num_tickets" binding:"required,min=1"`
}

func BookEvent(c *gin.Context) {
	eventID := c.Param("id")
	customerID := c.GetInt64("user_id")

	var req bookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx, err := db.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not start transaction"})
		return
	}
	defer tx.Rollback()

	// Atomic: decrement available_tickets only if enough remain
	res, err := tx.Exec(
		`UPDATE events SET available_tickets = available_tickets - ?
		 WHERE id = ? AND available_tickets >= ?`,
		req.NumTickets, eventID, req.NumTickets,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not update tickets"})
		return
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "not enough tickets available"})
		return
	}

	bookRes, err := tx.Exec(
		`INSERT INTO bookings (customer_id, event_id, num_tickets) VALUES (?, ?, ?)`,
		customerID, eventID, req.NumTickets,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create booking"})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not commit transaction"})
		return
	}

	bookingID, _ := bookRes.LastInsertId()

	// Fire confirmation job
	var customerEmail, eventName string
	db.DB.QueryRow(`SELECT email FROM users WHERE id = ?`, customerID).Scan(&customerEmail)
	db.DB.QueryRow(`SELECT name FROM events WHERE id = ?`, eventID).Scan(&eventName)

	worker.Enqueue(worker.Job{
		Type: worker.JobBookingConfirmation,
		Payload: map[string]any{
			"email":       customerEmail,
			"event_name":  eventName,
			"num_tickets": req.NumTickets,
		},
	})

	c.JSON(http.StatusCreated, gin.H{
		"booking_id":  bookingID,
		"event_id":    eventID,
		"num_tickets": req.NumTickets,
	})
}

func ListBookings(c *gin.Context) {
	customerID := c.GetInt64("user_id")

	rows, err := db.DB.Query(
		`SELECT id, customer_id, event_id, num_tickets, created_at FROM bookings WHERE customer_id = ? ORDER BY created_at DESC`,
		customerID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch bookings"})
		return
	}
	defer rows.Close()

	var bookings []models.Booking
	for rows.Next() {
		var b models.Booking
		rows.Scan(&b.ID, &b.CustomerID, &b.EventID, &b.NumTickets, &b.CreatedAt)
		bookings = append(bookings, b)
	}
	if bookings == nil {
		bookings = []models.Booking{}
	}
	c.JSON(http.StatusOK, bookings)
}

// ensure event exists helper used in BookEvent path
func eventExists(tx *sql.Tx, eventID string) bool {
	var id int64
	err := tx.QueryRow(`SELECT id FROM events WHERE id = ?`, eventID).Scan(&id)
	return err == nil
}
