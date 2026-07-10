package archtest

import (
	"go/types"
	"strings"
	"testing"
)

// TestPresentation_ConsumersFollowNamingConvention is the Go equivalent of
// PresentationTests.cs. Any type that subscribes to the in-process
// EventBus (i.e. declares a Handle method and lives in an app/consumers
// package) is an integration-event handler and must be named accordingly —
// mirroring the project's own existing convention
// (UserRegisteredConsumer, UserProfileUpdatedConsumer).
func TestPresentation_ConsumersFollowNamingConvention(t *testing.T) {
	pkgs := load(t)

	for _, p := range pkgs {
		if !strings.HasSuffix(p.PkgPath, "/internal/app/consumers") {
			continue
		}
		p := p

		t.Run(p.PkgPath, func(t *testing.T) {
			exportedNamedStructs(p, func(name string, named *types.Named, st *types.Struct) {
				if !implementsMethod(named, "Handle") {
					return
				}
				if !strings.HasSuffix(name, "Consumer") {
					fail(t,
						"%s.%s has a Handle method (subscribes to the event bus) but its name doesn't end in \"Consumer\"",
						p.PkgPath, name,
					)
				}
			})
		})
	}
}
