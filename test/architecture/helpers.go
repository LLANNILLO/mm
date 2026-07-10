// Package archtest enforces the modular-monolith, hexagonal and CQRS
// conventions of this codebase through static analysis of the package
// import graph and type declarations — the Go equivalent of NetArchTest
// in the C# reference project. It never imports application code directly;
// it loads and inspects it via go/packages, the same machinery gopls and
// go vet use.
package archtest

import (
	"fmt"
	"go/types"
	"strings"
	"sync"
	"testing"

	"golang.org/x/tools/go/packages"
)

const modulePrefix = "github.com/llannillo/mm/modules/"

const domainEventsPkgPath = "github.com/llannillo/mm/internal/shared/events"

var moduleNames = []string{"events", "ticketing", "users"}

var (
	loadOnce sync.Once
	loaded   []*packages.Package
	loadErr  error
)

// load parses and type-checks every package in the module exactly once per
// test binary run, then hands back the shared result to every test.
func load(t *testing.T) []*packages.Package {
	t.Helper()
	loadOnce.Do(func() {
		cfg := &packages.Config{
			Mode: packages.NeedName | packages.NeedFiles | packages.NeedImports |
				packages.NeedDeps | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedSyntax,
		}
		loaded, loadErr = packages.Load(cfg, "github.com/llannillo/mm/...")
	})
	if loadErr != nil {
		t.Fatalf("load packages: %v", loadErr)
	}
	for _, p := range loaded {
		for _, e := range p.Errors {
			t.Errorf("package %s failed to load: %v", p.PkgPath, e)
		}
	}
	return loaded
}

func packageByPath(pkgs []*packages.Package, path string) *packages.Package {
	for _, p := range pkgs {
		if p.PkgPath == path {
			return p
		}
	}
	return nil
}

func domainPackage(t *testing.T, pkgs []*packages.Package, module string) *packages.Package {
	t.Helper()
	path := modulePrefix + module + "/internal/domain"
	p := packageByPath(pkgs, path)
	if p == nil {
		t.Fatalf("domain package not found for module %q at %s", module, path)
	}
	return p
}

// domainEventInterface returns the events.DomainEvent interface type so
// callers can test other types against it with types.Implements.
func domainEventInterface(t *testing.T, pkgs []*packages.Package) *types.Interface {
	t.Helper()
	p := packageByPath(pkgs, domainEventsPkgPath)
	if p == nil {
		t.Fatalf("shared events package not found at %s", domainEventsPkgPath)
	}
	obj := p.Types.Scope().Lookup("DomainEvent")
	if obj == nil {
		t.Fatalf("DomainEvent interface not found in %s", domainEventsPkgPath)
	}
	iface, ok := obj.Type().Underlying().(*types.Interface)
	if !ok {
		t.Fatalf("%s.DomainEvent is not an interface", domainEventsPkgPath)
	}
	return iface
}

// exportedNamedStructs iterates every exported package-level type in pkg
// whose underlying type is a struct, invoking fn for each.
func exportedNamedStructs(pkg *packages.Package, fn func(name string, named *types.Named, st *types.Struct)) {
	scope := pkg.Types.Scope()
	for _, name := range scope.Names() {
		obj := scope.Lookup(name)
		tn, ok := obj.(*types.TypeName)
		if !ok || !tn.Exported() {
			continue
		}
		named, ok := obj.Type().(*types.Named)
		if !ok {
			continue
		}
		st, ok := named.Underlying().(*types.Struct)
		if !ok {
			continue
		}
		fn(name, named, st)
	}
}

// implementsMethod reports whether named (or *named) declares a method
// with the given name, regardless of receiver kind.
func implementsMethod(named *types.Named, methodName string) bool {
	for i := 0; i < named.NumMethods(); i++ {
		if named.Method(i).Name() == methodName {
			return true
		}
	}
	return false
}

// assertNoExportedFields fails t if st has any exported field. entityLike
// names the type for error messages.
func assertNoExportedFields(t *testing.T, pkgPath, typeName string, st *types.Struct) {
	t.Helper()
	for i := 0; i < st.NumFields(); i++ {
		f := st.Field(i)
		if f.Exported() {
			t.Errorf(
				"%s.%s has exported field %q — domain types must expose state only through methods, never raw fields, so nothing outside the package can construct an invalid instance",
				pkgPath, typeName, f.Name(),
			)
		}
	}
}

// assertValidateSignature fails t if named declares a Validate method whose
// signature is not exactly func() error.
func assertValidateSignature(t *testing.T, pkgPath, typeName string, named *types.Named) {
	t.Helper()
	for i := 0; i < named.NumMethods(); i++ {
		m := named.Method(i)
		if m.Name() != "Validate" {
			continue
		}
		sig, ok := m.Type().(*types.Signature)
		if !ok {
			continue
		}
		if sig.Params().Len() != 0 || sig.Results().Len() != 1 || sig.Results().At(0).Type().String() != "error" {
			t.Errorf(
				"%s.%s.Validate has signature %s — must be exactly func() error",
				pkgPath, typeName, sig.String(),
			)
		}
	}
}

func packagesUnder(pkgs []*packages.Package, prefix string) []*packages.Package {
	var out []*packages.Package
	for _, p := range pkgs {
		if strings.HasPrefix(p.PkgPath, prefix) {
			out = append(out, p)
		}
	}
	return out
}

func fail(t *testing.T, format string, args ...any) {
	t.Helper()
	t.Error(fmt.Sprintf(format, args...))
}
