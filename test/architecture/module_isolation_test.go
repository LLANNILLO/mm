package archtest

import (
	"strings"
	"testing"
)

// TestModuleIsolation_NoModuleDependsOnAnotherModule is the Go equivalent of
// ModuleTests.cs: a module may only ever depend on another module's
// integration events (its async, public contract). It must never depend on
// another module's internal packages, and never on another module's
// synchronous api package — that dependency style was removed when the
// project moved to the async in-process EventBus.
func TestModuleIsolation_NoModuleDependsOnAnotherModule(t *testing.T) {
	pkgs := load(t)

	for _, mod := range moduleNames {
		mod := mod
		t.Run(mod, func(t *testing.T) {
			modPkgs := packagesUnder(pkgs, modulePrefix+mod)
			for _, p := range modPkgs {
				for impPath := range p.Imports {
					for _, other := range moduleNames {
						if other == mod {
							continue
						}

						otherInternalPrefix := modulePrefix + other + "/internal"
						if strings.HasPrefix(impPath, otherInternalPrefix) {
							fail(t,
								"%s imports %s — a module must never depend on another module's internal packages",
								p.PkgPath, impPath,
							)
						}

						otherAPIPrefix := modulePrefix + other + "/api"
						otherIntegrationEvents := otherAPIPrefix + "/integrationevents"
						if strings.HasPrefix(impPath, otherAPIPrefix) && impPath != otherIntegrationEvents {
							fail(t,
								"%s imports %s — modules may only depend on another module's integration events (%s), never its synchronous api package",
								p.PkgPath, impPath, otherIntegrationEvents,
							)
						}
					}
				}
			}
		})
	}
}
