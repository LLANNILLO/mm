// Package outbox implements the transactional outbox and idempotent-consumer
// patterns shared by every module. Each module owns its own outbox_messages
// and outbox_message_consumers tables (same schema, one instance per module),
// wired together by the types in this package.
package outbox

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/llannillo/mm/internal/shared/events"
)

// TypeRegistry maps a domain event's type name to its concrete Go type, so the
// worker can reconstruct an events.DomainEvent from the JSON stored in
// outbox_messages.content. Equivalent to Newtonsoft's TypeNameHandling in the
// C# reference, but keyed by our own "type" column instead of embedding
// .NET-specific metadata in the payload.
type TypeRegistry struct {
	types map[string]reflect.Type
}

func NewTypeRegistry() *TypeRegistry {
	return &TypeRegistry{types: make(map[string]reflect.Type)}
}

// RegisterType makes T decodable by its type name.
func RegisterType[T events.DomainEvent](r *TypeRegistry) {
	var zero T
	r.types[reflect.TypeOf(zero).Name()] = reflect.TypeOf(zero)
}

// Decode reconstructs the concrete domain event named typeName from content.
func (r *TypeRegistry) Decode(typeName string, content []byte) (events.DomainEvent, error) {
	t, ok := r.types[typeName]
	if !ok {
		return nil, fmt.Errorf("outbox: unregistered domain event type %q", typeName)
	}

	ptr := reflect.New(t)
	if err := json.Unmarshal(content, ptr.Interface()); err != nil {
		return nil, fmt.Errorf("outbox: decode %q: %w", typeName, err)
	}

	domainEvent, ok := ptr.Elem().Interface().(events.DomainEvent)
	if !ok {
		return nil, fmt.Errorf("outbox: %q does not implement events.DomainEvent", typeName)
	}
	return domainEvent, nil
}
