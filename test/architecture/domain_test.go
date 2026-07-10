package archtest

import (
	"go/types"
	"strings"
	"testing"
)

// entityTypes lists, per module, the aggregate/entity types that must be
// fully encapsulated: zero exported fields, an exported New<Type> factory
// that enforces invariants, and an exported Rehydrate<Type> used only by
// the repository to reconstruct persisted state.
var entityTypes = map[string][]string{
	"events":    {"Category", "Event", "TicketType"},
	"ticketing": {"Customer", "Event", "TicketType", "Order", "Ticket", "Payment"},
	"users":     {"User"},
}

// childValueObjects lists types that belong to an aggregate but are only
// ever constructed by their parent (never directly, so no New<Type> is
// expected) — they still must be fully encapsulated and rehydratable.
var childValueObjects = map[string][]string{
	"ticketing": {"OrderItem"},
}

// TestDomain_EntitiesAreEncapsulated is the Go equivalent of
// DomainTests.Entities_ShouldHave_PrivateParameterlessConstructor and
// Entities_ShouldOnlyHave_PrivateConstructors. C# enforces "nobody can
// construct an invalid instance" via constructor visibility; Go has no
// constructors, so the equivalent is field visibility: every field must be
// unexported, forcing all construction through the package's own factory or
// rehydration functions.
func TestDomain_EntitiesAreEncapsulated(t *testing.T) {
	pkgs := load(t)

	for module, names := range entityTypes {
		module, names := module, names
		pkg := domainPackage(t, pkgs, module)
		for _, name := range names {
			name := name
			t.Run(module+"/"+name, func(t *testing.T) {
				obj := pkg.Types.Scope().Lookup(name)
				if obj == nil {
					t.Fatalf("type %s not found in %s", name, pkg.PkgPath)
				}
				named, ok := obj.Type().(*types.Named)
				if !ok {
					t.Fatalf("%s is not a named type", name)
				}
				st, ok := named.Underlying().(*types.Struct)
				if !ok {
					t.Fatalf("%s is not a struct", name)
				}

				assertNoExportedFields(t, pkg.PkgPath, name, st)

				if pkg.Types.Scope().Lookup("New"+name) == nil {
					fail(t, "%s: missing exported constructor New%s", pkg.PkgPath, name)
				}
				if pkg.Types.Scope().Lookup("Rehydrate"+name) == nil {
					fail(t, "%s: missing exported Rehydrate%s for repository reconstruction", pkg.PkgPath, name)
				}
			})
		}
	}

	for module, names := range childValueObjects {
		module, names := module, names
		pkg := domainPackage(t, pkgs, module)
		for _, name := range names {
			name := name
			t.Run(module+"/"+name+"(child)", func(t *testing.T) {
				obj := pkg.Types.Scope().Lookup(name)
				if obj == nil {
					t.Fatalf("type %s not found in %s", name, pkg.PkgPath)
				}
				named, ok := obj.Type().(*types.Named)
				if !ok {
					t.Fatalf("%s is not a named type", name)
				}
				st, ok := named.Underlying().(*types.Struct)
				if !ok {
					t.Fatalf("%s is not a struct", name)
				}

				assertNoExportedFields(t, pkg.PkgPath, name, st)

				if pkg.Types.Scope().Lookup("Rehydrate"+name) == nil {
					fail(t, "%s: missing exported Rehydrate%s for repository reconstruction", pkg.PkgPath, name)
				}
			})
		}
	}
}

// TestDomain_DomainEventsAreNamedConsistently is the Go equivalent of
// DomainTests.DomainEvent_ShouldHave_DomainEventPostfix. "Sealed" has no Go
// equivalent (structs cannot be subclassed at all, so that half of the C#
// rule is structurally impossible to violate).
func TestDomain_DomainEventsAreNamedConsistently(t *testing.T) {
	pkgs := load(t)
	domainEventIface := domainEventInterface(t, pkgs)

	for _, module := range moduleNames {
		module := module
		pkg := domainPackage(t, pkgs, module)
		t.Run(module, func(t *testing.T) {
			exportedNamedStructs(pkg, func(name string, named *types.Named, st *types.Struct) {
				if !types.Implements(named, domainEventIface) && !types.Implements(types.NewPointer(named), domainEventIface) {
					return
				}
				if !strings.HasSuffix(name, "DomainEvent") {
					fail(t,
						"%s.%s implements events.DomainEvent but its name doesn't end in \"DomainEvent\"",
						pkg.PkgPath, name,
					)
				}
			})
		})
	}
}
