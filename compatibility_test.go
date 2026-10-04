package dd_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"reflect"
	"strings"
	"testing"

	"github.com/Code-Hex/dd"
)

// Check generated expressions with the type checker, not just the parser.
// Local test types are declared in the synthetic package below.
func checkDumpType(t *testing.T, got, declarations, typ string) {
	t.Helper()
	source := "package fixture\n" + declarations + "\nvar _ " + typ + " = " + strings.ReplaceAll(got, "dd_test.", "")
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "dump.go", source, 0)
	if err != nil {
		t.Fatalf("parse dump: %v\n%s", err, source)
	}
	if _, err := new(types.Config).Check("fixture", fset, []*ast.File{file}, nil); err != nil {
		t.Fatalf("type-check dump: %v\n%s", err, source)
	}
}

type variadicFunc func(string, ...int) (int, error)
type resultFunc func() string

func TestNamedFunctionSignatures(t *testing.T) {
	cases := []struct {
		value     interface{}
		decl, typ string
	}{
		{variadicFunc(func(string, ...int) (int, error) { panic("must not be called") }), "type variadicFunc func(string, ...int) (int, error)", "variadicFunc"},
		{resultFunc(func() string { panic("must not be called") }), "type resultFunc func() string", "resultFunc"},
		{variadicFunc(nil), "type variadicFunc func(string, ...int) (int, error)", "variadicFunc"},
	}
	for _, tc := range cases {
		t.Run(tc.typ+reflect.ValueOf(tc.value).String(), func(t *testing.T) {
			checkDumpType(t, dd.Dump(tc.value), tc.decl, tc.typ)
		})
	}
}

func TestReflectValueAsData(t *testing.T) {
	// Validate the public behavior without fixing reflect.Value's private layout.
	got := dd.Dump(reflect.Value{})
	if !strings.HasPrefix(got, "reflect.Value{") {
		t.Fatalf("reflect.Value was unwrapped: %s", got)
	}
	if _, err := parser.ParseExpr(got); err != nil {
		t.Fatal(err)
	}
	if got := dd.Dump(reflect.Value{}, dd.WithExportedOnly()); got != "reflect.Value{}" {
		t.Fatal(got)
	}
}
