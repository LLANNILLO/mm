package domain

type entity struct {
	domainEvents []DomainEvent
}

func (e *entity) raise(event DomainEvent) {
	e.domainEvents = append(e.domainEvents, event)
}

func (e *entity) DomainEvents() []DomainEvent {
	return e.domainEvents
}

func (e *entity) ClearDomainEvents() {
	e.domainEvents = nil
}
