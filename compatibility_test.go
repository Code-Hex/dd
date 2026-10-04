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

func TestFunctionZeroFormatting(t *testing.T) {
	t.Run("custom primitive", func(t *testing.T) {
		calls := 0
		got := dd.Dump(func() int { return 0 }, dd.WithDumpFunc(func(n int, w dd.Writer) {
			if n != 0 {
				t.Fatalf("expected zero, got %d", n)
			}
			calls++
			w.Write("123")
		}))
		if !strings.Contains(got, "return 123") || calls != 1 {
			t.Fatalf("calls=%d dump=%s", calls, got)
		}
	})
	t.Run("hex", func(t *testing.T) {
		got := dd.Dump(func() uint8 { return 0 }, dd.WithUintFormat(dd.HexUint))
		if !strings.Contains(got, "return 0x00") {
			t.Fatal(got)
		}
	})
	t.Run("nested result", func(t *testing.T) {
		type result struct{ N int }
		f := func() result { return result{} }
		value := struct {
			A func() result
			B struct{ F func() result }
		}{A: f, B: struct{ F func() result }{F: f}}
		got := dd.Dump(value)
		if !strings.Contains(got, "return dd_test.result{\n        N: 0,\n      }") {
			t.Fatal(got)
		}
	})
}

func TestPointerCustomDump(t *testing.T) {
	type result struct{ N int }
	value := &result{N: 42}
	got := dd.Dump(value, dd.WithOmitEmptyFields(), dd.WithDumpFunc(func(v result, w dd.Writer) {
		if v != *value {
			t.Fatalf("wrong value: %v", v)
		}
		w.Write("dd_test.result{N: 42}")
	}))
	if got != "&dd_test.result{N: 42}" {
		t.Fatal(got)
	}
	checkDumpType(t, got, "type result struct { N int }", "*result")
}

func TestFunctionZeroCustomRepeatedResults(t *testing.T) {
	type result struct{ N int }
	calls := 0
	got := dd.Dump(func() (result, result) { return result{}, result{} },
		dd.WithDumpFunc(func(v result, w dd.Writer) {
			calls++
			w.Write("dd_test.result{}")
		}))
	if calls != 2 {
		t.Fatalf("expected each result to use the callback, calls=%d dump=%s", calls, got)
	}
	checkDumpType(t, got, "type result struct { N int }", "func() (result, result)")
}
