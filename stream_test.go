package dd_test

import (
	"bytes"
	"errors"
	"io"
	"math/rand"
	"strconv"
	"strings"
	"testing"

	"github.com/Code-Hex/dd"
)

type streamWriter struct {
	bytes.Buffer
	err       error
	failAfter int
	writes    int
}

func (w *streamWriter) Write(p []byte) (int, error) {
	w.writes++
	if w.writes > w.failAfter {
		return 0, w.err
	}
	return w.Buffer.Write(p)
}

// Hide Buffer.WriteString so every write exercises the same failing surface.
func (w *streamWriter) WriteString(s string) (int, error) { return w.Write([]byte(s)) }

func TestDumpToErrors(t *testing.T) {
	sentinel := errors.New("write failed")
	for _, value := range []interface{}{[]int{1, 2, 3}, map[string]int{"x": 1, "y": 2}} {
		for _, err := range []error{sentinel, nil} {
			w := &streamWriter{err: err, failAfter: 1}
			got := dd.DumpTo(w, value)
			want := err
			if want == nil {
				want = io.ErrShortWrite
			}
			if !errors.Is(got, want) {
				t.Fatalf("want %v, got %v", want, got)
			}
		}
	}
}

func TestDumpToStreamsAndStops(t *testing.T) {
	w := &streamWriter{failAfter: 100000}
	calls := 0
	value := make([]int, 1000)
	opt := dd.WithDumpFunc(func(v int, out dd.Writer) {
		if w.Len() == 0 {
			t.Fatal("output buffered before callback")
		}
		calls++
		out.Write("1")
	})
	if err := dd.DumpTo(w, value, opt); err != nil {
		t.Fatal(err)
	}
	if calls != len(value) || strings.Count(w.String(), "1,") != len(value) {
		t.Fatal("incomplete output")
	}

	w = &streamWriter{err: errors.New("stop"), failAfter: 0}
	calls = 0
	if dd.DumpTo(w, value, opt) == nil || calls != 0 {
		t.Fatalf("callbacks continued after failure: %d", calls)
	}
}

func TestDumpToFormatting(t *testing.T) {
	type key struct{ N int }
	cases := []struct {
		value interface{}
		want  string
	}{
		{map[key]interface{}{{1}: 42, {22}: "hello"},
			"map[dd_test.key]interface {}{\n  dd_test.key{\n    N: 1,\n  }: 42,\n     dd_test.key{\n    N: 22,\n  }: \"hello\",\n}"},
		{map[string]int{"あ": 1, "long": 2},
			"map[string]int{\n  \"long\": 2,\n  \"あ\":    1,\n}"},
	}
	for _, tc := range cases {
		var out bytes.Buffer
		if err := dd.DumpTo(&out, tc.value); err != nil {
			t.Fatal(err)
		}
		if out.String() != tc.want {
			t.Fatalf("want %q, got %q", tc.want, out.String())
		}
		if dd.Dump(tc.value) != tc.want {
			t.Fatal("Dump wrapper changed output")
		}
	}
}

func TestDumpSharedAddresses(t *testing.T) {
	x := make([]interface{}, 1)
	x[0] = x[:0]
	if got := dd.Dump(x); strings.Contains(got, "unsafe.Pointer") {
		t.Fatal(got)
	}
	type inner struct{ N int }
	type outer struct {
		inner
		P *inner
	}
	value := &outer{inner: inner{N: 1}}
	value.P = &value.inner
	if got := dd.Dump(value); strings.Contains(got, "unsafe.Pointer") || strings.Count(got, "N: 1,") != 2 {
		t.Fatal(got)
	}
	x[0] = x
	if got := dd.Dump(x); !strings.Contains(got, "unsafe.Pointer") {
		t.Fatal("cycle was not detected")
	}
}

func TestDumpToLargeCustomBlock(t *testing.T) {
	line := strings.Repeat("x", 128*1024)
	var out bytes.Buffer
	err := dd.DumpTo(&out, 0, dd.WithDumpFunc(func(v int, w dd.Writer) { w.WriteBlock(line + "\nend") }))
	if err != nil {
		t.Fatal(err)
	}
	if want := "{\n  " + line + "\n  end\n}"; out.String() != want {
		t.Fatalf("large block truncated: got %d bytes", out.Len())
	}
}

func BenchmarkDumpTo(b *testing.B) {
	values := make([]benchmarkRecord, 10000)
	for i := range values {
		values[i] = benchmarkRecord{i, "example", true, []int{1, 2, 3}}
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := dd.DumpTo(io.Discard, values); err != nil {
			b.Fatal(err)
		}
	}
}

type byteCounter int64

func (n *byteCounter) Write(p []byte) (int, error) { *n += byteCounter(len(p)); return len(p), nil }
func (n *byteCounter) WriteString(s string) (int, error) {
	*n += byteCounter(len(s))
	return len(s), nil
}

func TestDumpToLargeValues(t *testing.T) {
	var n byteCounter
	values := make([]int, 1000000)
	if err := dd.DumpTo(&n, values); err != nil {
		t.Fatal(err)
	}
	// Header, one million "  0,\n" lines, and closing brace.
	if want := byteCounter(len("[]int{\n}") + len(values)*5); n != want {
		t.Fatalf("want %d bytes, got %d", want, n)
	}
	var node *benchmarkNode
	for i := 0; i < 2048; i++ {
		node = &benchmarkNode{Next: node}
	}
	n = 0
	if err := dd.DumpTo(&n, node); err != nil {
		t.Fatal(err)
	}
	const depth = 2048
	want := byteCounter(3*depth*depth + depth +
		depth*(len("&dd_test.benchmarkNode{\n")+len("Value: 0,\n")+len("Next: ")+len(",\n}")) +
		len("(*dd_test.benchmarkNode)(nil)"))
	if n != want {
		t.Fatalf("want %d bytes, got %d", want, n)
	}
}

func TestCustomBlockTabs(t *testing.T) {
	got := dd.Dump(0, dd.WithIndent(4), dd.WithDumpFunc(func(v int, w dd.Writer) { w.WriteBlock("a\tb") }))
	if want := "{\n    a   b\n}"; got != want {
		t.Fatalf("want %q, got %q", want, got)
	}
}

func TestStreamQuotedStrings(t *testing.T) {
	cases := []string{"", "quote\" slash\\ newline\n", strings.Repeat("a", 4095) + "世界", strings.Repeat("\x80", 8193), strings.Repeat("é", 4096), strings.Repeat("\xff\x00\u2028", 2048)}
	random := rand.New(rand.NewSource(1))
	for _, size := range []int{4095, 4096, 4097, 12000} {
		data := make([]byte, size)
		random.Read(data)
		cases = append(cases, string(data))
	}
	for _, value := range cases {
		var out bytes.Buffer
		if err := dd.DumpTo(&out, value); err != nil {
			t.Fatal(err)
		}
		if want := strconv.Quote(value); out.String() != want {
			t.Fatalf("quoted output differs for %d bytes", len(value))
		}
	}
}

type failPlainWriter struct {
	calls int
	err   error
}

func (w *failPlainWriter) Write(p []byte) (int, error) { w.calls++; return len(p) / 2, w.err }

func TestAllWritePathsStop(t *testing.T) {
	sentinel := errors.New("destination failed")
	for _, value := range []interface{}{1, 1.5, strings.Repeat("é", 8192), map[string]int{"a": 1}} {
		for _, failure := range []error{nil, sentinel} {
			w := &failPlainWriter{err: failure}
			got := dd.DumpTo(w, value)
			want := failure
			if want == nil {
				want = io.ErrShortWrite
			}
			if got != want || w.calls != 1 {
				t.Fatalf("got %v after %d calls; want %v after one call", got, w.calls, want)
			}
		}
	}
}

func TestPlainWriterOutput(t *testing.T) {
	var out bytes.Buffer
	// Expose only Write to exercise the bounded string-to-byte adapter.
	dst := struct{ io.Writer }{&out}
	value := strings.Repeat("a", 16384)
	if err := dd.DumpTo(dst, value); err != nil {
		t.Fatal(err)
	}
	if out.String() != strconv.Quote(value) {
		t.Fatal("plain writer output differs")
	}
}

type delayedPlainWriter struct {
	successes, calls int
	err              error
}

func (w *delayedPlainWriter) Write(p []byte) (int, error) {
	w.calls++
	if w.calls > w.successes {
		return len(p) / 2, w.err
	}
	return len(p), nil
}
func TestWriteFailuresAfterProgress(t *testing.T) {
	sentinel := errors.New("late failure")
	for _, value := range []interface{}{strings.Repeat("x", 20000), []float64{1.5, 2.5, 3.5}, map[string]int{"a": 1, "b": 2}} {
		for _, successes := range []int{1, 2} {
			for _, failure := range []error{nil, sentinel} {
				w := &delayedPlainWriter{successes: successes, err: failure}
				got := dd.DumpTo(w, value)
				want := failure
				if want == nil {
					want = io.ErrShortWrite
				}
				if got != want || w.calls != successes+1 {
					t.Fatalf("%T: got %v, calls=%d", value, got, w.calls)
				}
			}
		}
	}
}
