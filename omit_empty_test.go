package dd_test

import (
	"strings"
	"testing"

	"github.com/Code-Hex/dd"
)

func TestOmitEmptyFields(t *testing.T) {
	type sample struct {
		Bool      bool
		Number    int
		String    string
		Array     [2]int
		Struct    struct{ N int }
		Slice     []int
		Map       map[string]int
		Pointer   *int
		Interface interface{}
		Func      func()
		Chan      chan int
		hidden    int
	}
	if got := dd.Dump(sample{}, dd.WithOmitEmptyFields()); got != "dd_test.sample{}" {
		t.Fatal(got)
	}
	if got := dd.Dump(sample{}); !strings.Contains(got, "Number: 0,") {
		t.Fatalf("default output omitted fields: %s", got)
	}
	value := sample{Number: 2, Slice: []int{}, Map: map[string]int{}, Interface: (*int)(nil), hidden: 3}
	want := "dd_test.sample{\n  Number: 2,\n  Slice: []int{},\n  Map: map[string]int{},\n  Interface: (*int)(nil),\n  hidden: 3,\n}"
	if got := dd.Dump(value, dd.WithOmitEmptyFields()); got != want {
		t.Fatalf("want %q, got %q", want, got)
	}
	got := dd.Dump(&value, dd.WithOmitEmptyFields(), dd.WithExportedOnly())
	if strings.Contains(got, "hidden:") || !strings.Contains(got, "Number: 2,") {
		t.Fatal(got)
	}
	n := 0
	value.Pointer = &n
	value.Func = func() {}
	value.Chan = make(chan int)
	got = dd.Dump(value, dd.WithOmitEmptyFields())
	for _, field := range []string{"Pointer:", "Func:", "Chan:"} {
		if !strings.Contains(got, field) {
			t.Fatalf("missing nonzero %s in %s", field, got)
		}
	}
	for _, value := range []interface{}{
		[]sample{{Number: 2}},
		map[string]sample{"x": {Number: 2}},
		struct{ Child sample }{sample{Number: 2}},
	} {
		got := dd.Dump(value, dd.WithOmitEmptyFields(), dd.WithIndent(4))
		if strings.Contains(got, "Bool:") || !strings.Contains(got, "Number: 2,") {
			t.Fatal(got)
		}
	}
}

func TestZeroValueCacheIsolation(t *testing.T) {
	type result struct{ hidden int }
	f := func() result { panic("must not be called") }
	// Distinct calls must not inherit formatting or omission from previous calls.
	for _, opts := range [][]dd.OptionFunc{
		{dd.WithOmitEmptyFields()}, {dd.WithExportedOnly()}, nil,
	} {
		got := dd.Dump(f, opts...)
		if strings.Contains(got, "hidden:") != (opts == nil) {
			t.Fatal(got)
		}
	}
	for i := 0; i < 16; i++ {
		t.Run("parallel", func(t *testing.T) {
			t.Parallel()
			dd.Dump(f)
		})
	}
}

func TestOmitEmptyFieldsCustomDump(t *testing.T) {
	type custom struct{ Zero, Value int }
	calls := 0
	got := dd.Dump(custom{Value: 1}, dd.WithOmitEmptyFields(),
		dd.WithDumpFunc(func(n int, w dd.Writer) {
			calls++
			w.Write("99")
		}))
	if want := "dd_test.custom{\n  Value: 99,\n}"; got != want || calls != 1 {
		t.Fatalf("got %q, custom dumper called %d times", got, calls)
	}
}
