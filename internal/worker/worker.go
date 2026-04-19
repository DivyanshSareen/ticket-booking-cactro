package worker

import (
	"fmt"
	"log"
)

type JobType string

const (
	JobBookingConfirmation   JobType = "booking_confirmation"
	JobEventUpdateNotify     JobType = "event_update_notify"
)

type Job struct {
	Type    JobType
	Payload map[string]any
}

var jobQueue chan Job

func Start(numWorkers int) {
	jobQueue = make(chan Job, 100)
	for i := 0; i < numWorkers; i++ {
		go worker(i)
	}
	log.Printf("[worker] started %d workers", numWorkers)
}

func Enqueue(j Job) {
	jobQueue <- j
}

func worker(id int) {
	for job := range jobQueue {
		switch job.Type {
		case JobBookingConfirmation:
			email := job.Payload["email"]
			eventName := job.Payload["event_name"]
			numTickets := job.Payload["num_tickets"]
			fmt.Printf("[worker-%d] [EMAIL] Booking confirmation sent to %s — Event: %s, Tickets: %v\n",
				id, email, eventName, numTickets)

		case JobEventUpdateNotify:
			eventName := job.Payload["event_name"]
			customers := job.Payload["customers"]
			fmt.Printf("[worker-%d] [NOTIFY] Event '%s' updated — notifying customers: %v\n",
				id, eventName, customers)
		}
	}
}
