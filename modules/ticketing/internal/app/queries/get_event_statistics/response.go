package geteventstatistics

import "github.com/google/uuid"

type Response struct {
	EventID                 uuid.UUID `json:"event_id"`
	TicketsSold             int32     `json:"tickets_sold"`
	AttendeesCheckedIn      int32     `json:"attendees_checked_in"`
	DuplicateCheckInTickets []string  `json:"duplicate_check_in_tickets"`
	InvalidCheckInTickets   []string  `json:"invalid_check_in_tickets"`
}
