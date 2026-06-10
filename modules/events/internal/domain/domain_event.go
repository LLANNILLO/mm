package domain

import "github.com/google/uuid"

type DomainEvent interface {
	domainEvent()
}

type EventCreatedDomainEvent struct {
	EventID uuid.UUID
}

func (EventCreatedDomainEvent) domainEvent() {}
