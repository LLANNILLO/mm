package archtest

import (
	"go/types"
	"strings"
	"testing"
)

// TestApplication_CommandsAndQueriesFollowCQRSConventions is the Go
// equivalent of ApplicationTests.cs. C# names the DTO CreateEventCommand
// because the whole Application assembly shares one namespace; this
// codebase instead gives every command/query its own package
// (createevent.Command), so the naming rule becomes structural: if a
// Handler's Handle method takes an input parameter beyond context.Context,
// that parameter's type must be named exactly "Command"/"Query" and be
// declared in the same package. A parameterless query (e.g. "list all X")
// is allowed to skip the DTO entirely — unlike C#'s MediatR, nothing here
// requires an empty marker object to dispatch a query. Handler itself must
// always exist, be exported, have unexported fields (nothing outside the
// package may construct one except through NewHandler), and expose an
// exported NewHandler constructor.
func TestApplication_CommandsAndQueriesFollowCQRSConventions(t *testing.T) {
	pkgs := load(t)

	for _, p := range pkgs {
		isCommand := strings.Contains(p.PkgPath, "/internal/app/commands/")
		isQuery := strings.Contains(p.PkgPath, "/internal/app/queries/")
		if !isCommand && !isQuery {
			continue
		}
		p := p

		wantDTOName := "Query"
		if isCommand {
			wantDTOName = "Command"
		}

		t.Run(p.PkgPath, func(t *testing.T) {
			scope := p.Types.Scope()

			handlerObj := scope.Lookup("Handler")
			if handlerObj == nil {
				t.Fatalf("package %s must export a type named \"Handler\"", p.PkgPath)
			}
			handlerNamed, ok := handlerObj.Type().(*types.Named)
			if !ok {
				t.Fatalf("package %s: Handler is not a named type", p.PkgPath)
			}
			if st, ok := handlerNamed.Underlying().(*types.Struct); ok {
				assertNoExportedFields(t, p.PkgPath, "Handler", st)
			}
			if scope.Lookup("NewHandler") == nil {
				fail(t, "package %s: Handler must have an exported NewHandler constructor", p.PkgPath)
			}

			handleParamType := handleInputType(handlerNamed)
			if handleParamType == nil {
				// Parameterless query/command (e.g. "list all X") — no DTO required.
				return
			}
			named, ok := handleParamType.(*types.Named)
			if !ok || named.Obj().Pkg() != p.Types || named.Obj().Name() != wantDTOName {
				fail(t,
					"package %s: Handler.Handle's input parameter must be named %q and declared in this package, got %s",
					p.PkgPath, wantDTOName, handleParamType.String(),
				)
				return
			}
			assertValidateSignature(t, p.PkgPath, wantDTOName, named)
		})
	}
}

// handleInputType returns the type of Handle's parameter after context.Context,
// or nil if Handle takes no input beyond the context.
func handleInputType(handler *types.Named) types.Type {
	for i := 0; i < handler.NumMethods(); i++ {
		m := handler.Method(i)
		if m.Name() != "Handle" {
			continue
		}
		sig, ok := m.Type().(*types.Signature)
		if !ok || sig.Params().Len() < 2 {
			return nil
		}
		return sig.Params().At(1).Type()
	}
	return nil
}
