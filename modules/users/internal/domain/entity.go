package domain

import "github.com/llannillo/mm/internal/shared/events"

type entity struct {
	domainEvents []events.DomainEvent
}

func (e *entity) raise(event events.DomainEvent) {
	e.domainEvents = append(e.domainEvents, event)
}

func (e *entity) DomainEvents() []events.DomainEvent {
	return e.domainEvents
}

func (e *entity) ClearDomainEvents() {
	e.domainEvents = nil
}
