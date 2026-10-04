package p_test

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/Code-Hex/dd"
	"github.com/Code-Hex/dd/p"
	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters"
)

type limitedWriter struct {
	buf   bytes.Buffer
	limit int
	err   error
	calls int
}

func (w *limitedWriter) Write(b []byte) (int, error) {
	w.calls++
	if len(b) > w.limit {
		b = b[:w.limit]
	}
	n, _ := w.buf.Write(b)
	w.limit -= n
	return n, w.err
}

func TestFpFormatterError(t *testing.T) {
	wantErr := errors.New("format failed")
	calls := 0
	dumps := 0
	dumpOpt := dd.WithDumpFunc(func(v int, w dd.Writer) { dumps++; w.Write("1") })
	formatter := chroma.FormatterFunc(func(w io.Writer, _ *chroma.Style, _ chroma.Iterator) error {
		calls++
		io.WriteString(w, "partial")
		return wantErr
	})
	var out bytes.Buffer
	n, err := p.New(p.WithFormatter(formatter), p.WithDumpOptions(dumpOpt)).Fp(&out, 1, 2)
	if n != 7 || err != wantErr || out.String() != "partial" || calls != 1 || dumps != 1 {
		t.Fatalf("n=%d err=%v output=%q calls=%d dumps=%d", n, err, out.String(), calls, dumps)
	}
}

func TestFpWriteErrors(t *testing.T) {
	wantErr := errors.New("write failed")
	for _, tc := range []struct {
		name              string
		limit             int
		writeErr, wantErr error
	}{
		{"partial", 3, wantErr, wantErr},
		{"short", 3, nil, io.ErrShortWrite},
		{"zero", 0, nil, io.ErrShortWrite},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out := &limitedWriter{limit: tc.limit, err: tc.writeErr}
			calls := 0
			formatter := chroma.FormatterFunc(func(w io.Writer, _ *chroma.Style, _ chroma.Iterator) error {
				calls++
				io.WriteString(w, "abcdef")
				io.WriteString(w, "ignored")
				return errors.New("later format error")
			})
			n, err := p.New(p.WithFormatter(formatter)).Fp(out, 1, 2)
			if n != tc.limit || err != tc.wantErr || out.buf.String() != "abcdef"[:tc.limit] || out.calls != 1 || calls != 1 {
				t.Fatalf("n=%d err=%v output=%q writes=%d formats=%d", n, err, out.buf.String(), out.calls, calls)
			}
		})
	}
}

func TestFpDefaultFormatterWriteError(t *testing.T) {
	wantErr := errors.New("write failed")
	out := &limitedWriter{limit: 2, err: wantErr}
	n, err := p.Fp(out, "hello", "world")
	if n != 2 || err != wantErr || out.calls != 1 {
		t.Fatalf("n=%d err=%v calls=%d", n, err, out.calls)
	}
}

func TestFpOutput(t *testing.T) {
	opts := []dd.OptionFunc{dd.WithIndent(2), dd.WithOmitEmptyFields()}
	printer := p.New(p.WithFormatter(formatters.NoOp), p.WithDumpOptions(opts...))
	data := struct {
		Value string
		Empty string
	}{Value: "日本語"}
	for _, args := range [][]interface{}{nil, {data}, {data, "tail"}} {
		var parts []string
		for _, a := range args {
			parts = append(parts, dd.Dump(a, opts...))
		}
		want := strings.Join(parts, " ") + "\n"
		var out bytes.Buffer
		n, err := printer.Fp(&out, args...)
		if err != nil || n != len(want) || out.String() != want {
			t.Fatalf("n=%d err=%v output=%q want=%q", n, err, out.String(), want)
		}
	}
}

func BenchmarkFpLarge(b *testing.B) {
	data := strings.Repeat("x", 1<<20)
	printer := p.New()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := printer.Fp(io.Discard, data); err != nil {
			b.Fatal(err)
		}
	}
}

func TestFpSeparatorAndNewlineErrors(t *testing.T) {
	for _, tc := range []struct {
		name  string
		args  []interface{}
		limit int
		want  string
	}{
		{"separator", []interface{}{1, 2}, 1, "1"},
		{"newline", []interface{}{1}, 1, "1"},
		{"empty", nil, 0, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out := &limitedWriter{limit: tc.limit}
			n, err := p.New(p.WithFormatter(formatters.NoOp)).Fp(out, tc.args...)
			if n != len(tc.want) || err != io.ErrShortWrite || out.buf.String() != tc.want {
				t.Fatalf("n=%d err=%v output=%q", n, err, out.buf.String())
			}
		})
	}
}
