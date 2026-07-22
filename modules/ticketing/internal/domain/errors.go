package domain

import "net/http"

type ErrorKind int

const (
	KindValidation ErrorKind = iota
	KindNotFound
	KindConflict
	KindProblem
)

type DomainError struct {
	Code    string
	Message string
	Kind    ErrorKind
}

func (e *DomainError) Error() string { return e.Message }

func (e *DomainError) StatusCode() int {
	switch e.Kind {
	case KindNotFound:
		return http.StatusNotFound
	case KindConflict:
		return http.StatusConflict
	case KindValidation, KindProblem:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

var (
	ErrCustomerNotFound   = &DomainError{Code: "customer.not_found", Message: "customer not found", Kind: KindNotFound}
	ErrTicketTypeNotFound = &DomainError{Code: "ticket_type.not_found", Message: "ticket type not found", Kind: KindNotFound}

	ErrEventNotFound         = &DomainError{Code: "event.not_found", Message: "event not found", Kind: KindNotFound}
	ErrEventAlreadyCancelled = &DomainError{Code: "event.already_cancelled", Message: "event already cancelled", Kind: KindConflict}

	ErrTicketTypeInsufficientQuantity = &DomainError{Code: "ticket_type.insufficient_quantity", Message: "insufficient ticket quantity available", Kind: KindValidation}

	ErrOrderNotFound             = &DomainError{Code: "order.not_found", Message: "order not found", Kind: KindNotFound}
	ErrOrderTicketsAlreadyIssued = &DomainError{Code: "order.tickets_already_issued", Message: "tickets have already been issued for this order", Kind: KindConflict}

	ErrTicketAlreadyArchived   = &DomainError{Code: "ticket.already_archived", Message: "ticket is already archived", Kind: KindConflict}
	ErrTicketCheckInInvalid    = &DomainError{Code: "ticket.check_in_invalid", Message: "ticket does not belong to this customer", Kind: KindValidation}
	ErrTicketAlreadyCheckedIn  = &DomainError{Code: "ticket.already_checked_in", Message: "ticket has already been checked in", Kind: KindConflict}
	ErrEventStatisticsNotFound = &DomainError{Code: "event_statistics.not_found", Message: "no statistics found for this event", Kind: KindNotFound}

	ErrPaymentAlreadyRefunded     = &DomainError{Code: "payment.already_refunded", Message: "payment has already been fully refunded", Kind: KindConflict}
	ErrPaymentRefundExceedsAmount = &DomainError{Code: "payment.refund_exceeds_amount", Message: "refund amount exceeds payment amount", Kind: KindValidation}
)
