package main

import (
	"log"
	"os"

	"ticket-booking/internal/db"
	"ticket-booking/internal/handlers"
	"ticket-booking/internal/middleware"
	"ticket-booking/internal/worker"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, using environment variables")
	}

	db.Init("./ticket-booking.db")
	worker.Start(3)

	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.String(200, "Ticket Booking API — see /health for status")
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Public
	r.POST("/auth/register", handlers.Register)
	r.POST("/auth/login", handlers.Login)

	// Authenticated (both roles)
	auth := r.Group("/", middleware.Auth())
	auth.GET("/events", handlers.ListEvents)
	auth.GET("/events/:id", handlers.GetEvent)

	// Organizer only
	organizer := r.Group("/", middleware.Auth(), middleware.RequireRole("organizer"))
	organizer.POST("/events", handlers.CreateEvent)
	organizer.PUT("/events/:id", handlers.UpdateEvent)

	// Customer only
	customer := r.Group("/", middleware.Auth(), middleware.RequireRole("customer"))
	customer.POST("/events/:id/book", handlers.BookEvent)
	customer.GET("/bookings", handlers.ListBookings)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server running on :%s", port)
	r.Run(":" + port)
}
