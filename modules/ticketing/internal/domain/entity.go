package domain

import sharedevents "github.com/llannillo/mm/internal/shared/events"

type entity struct {
	domainEvents []sharedevents.DomainEvent
}

func (e *entity) raise(event sharedevents.DomainEvent) {
	e.domainEvents = append(e.domainEvents, event)
}

func (e *entity) DomainEvents() []sharedevents.DomainEvent {
	return e.domainEvents
}

func (e *entity) ClearDomainEvents() {
	e.domainEvents = nil
}
