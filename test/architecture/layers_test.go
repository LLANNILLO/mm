package archtest

import (
	"strings"
	"testing"
)

// TestLayers_DependenciesPointInward is the Go equivalent of LayerTests.cs.
// It freezes the hexagonal dependency direction for every module: domain has
// zero internal dependencies, application depends only on domain and ports
// (never concrete adapters), and the driving/driven adapter sides never
// depend on each other directly. The Go compiler already forbids any
// cross-*module* violation of this shape via the internal/ visibility rule;
// this test guards the same discipline *within* a single module, which the
// compiler cannot see.
func TestLayers_DependenciesPointInward(t *testing.T) {
	pkgs := load(t)

	type rule struct {
		from      string
		forbidden []string
	}
	rules := []rule{
		{from: "domain", forbidden: []string{"internal/app", "internal/ports", "internal/adapters"}},
		{from: "internal/app", forbidden: []string{"internal/adapters"}},
		{from: "internal/ports", forbidden: []string{"internal/adapters"}},
		{from: "internal/adapters/driving", forbidden: []string{"internal/adapters/driven"}},
		{from: "internal/adapters/driven", forbidden: []string{"internal/adapters/driving"}},
	}

	for _, mod := range moduleNames {
		mod := mod
		for _, r := range rules {
			r := r
			t.Run(mod+"/"+r.from, func(t *testing.T) {
				fromPrefix := modulePrefix + mod + "/" + r.from
				for _, p := range packagesUnder(pkgs, fromPrefix) {
					for impPath := range p.Imports {
						for _, forbidden := range r.forbidden {
							forbiddenPrefix := modulePrefix + mod + "/" + forbidden
							if strings.HasPrefix(impPath, forbiddenPrefix) {
								fail(t,
									"%s (layer %q) imports %s (layer %q) — dependencies must point inward only",
									p.PkgPath, r.from, impPath, forbidden,
								)
							}
						}
					}
				}
			})
		}
	}
}

// concreteInfraPackages are third-party driver/client packages that belong
// exclusively to adapters/driven — never to app or domain. A local
// "adapters" path check (above) cannot see this kind of leak, because the
// dependency comes from an external module, not a local package: this is
// exactly how get_user_permissions ended up injecting *pgxpool.Pool and raw
// SQL straight into the Application layer.
var concreteInfraPackages = []string{
	"github.com/jackc/pgx/v5",
	"github.com/jackc/pgx/v5/pgxpool",
	"github.com/jackc/pgx/v5/pgtype",
	"database/sql",
	"github.com/valkey-io/valkey-go",
}

func TestLayers_ApplicationNeverImportsConcreteInfrastructure(t *testing.T) {
	pkgs := load(t)

	for _, mod := range moduleNames {
		mod := mod
		t.Run(mod, func(t *testing.T) {
			appPrefix := modulePrefix + mod + "/internal/app"
			domainPrefix := modulePrefix + mod + "/internal/domain"
			for _, p := range pkgs {
				if !strings.HasPrefix(p.PkgPath, appPrefix) && !strings.HasPrefix(p.PkgPath, domainPrefix) {
					continue
				}
				for impPath := range p.Imports {
					for _, infra := range concreteInfraPackages {
						if impPath == infra {
							fail(t,
								"%s imports %s directly — application/domain code must depend on a port interface, never on a concrete driver; the driver belongs in adapters/driven",
								p.PkgPath, infra,
							)
						}
					}
				}
			}
		})
	}
}
