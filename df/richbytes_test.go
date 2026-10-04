package df_test

import (
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/Code-Hex/dd"
	"github.com/Code-Hex/dd/df"
)

func BenchmarkRichBytes(b *testing.B) {
	data := make([]byte, 1<<20)
	for i := range data {
		data[i] = byte(i)
	}
	b.ReportAllocs()
	b.SetBytes(int64(len(data)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := dd.DumpTo(ioutil.Discard, data, df.WithRichBytes()); err != nil {
			b.Fatal(err)
		}
	}
}

func TestRichBytesRoundTrip(t *testing.T) {
	allBytes := make([]byte, 256)
	large := make([]byte, 65537)
	for i := range allBytes {
		allBytes[i] = byte(i)
	}
	for i := range large {
		large[i] = byte(i)
	}
	cases := [][]byte{nil, {}, allBytes[:1], allBytes[:15], allBytes[:16], allBytes[:17], allBytes, large}
	var source strings.Builder
	source.WriteString("package main\nimport (\"bytes\"; \"encoding/base64\"; \"fmt\")\nfunc main() {\n")
	for i, data := range cases {
		got := dd.Dump(data, df.WithRichBytes())
		if data != nil {
			lines := strings.Split(hex.Dump(data), "\n")
			for j, line := range lines {
				if line != "" {
					lines[j] = "// " + line
				}
			}
			var literal strings.Builder
			literal.WriteString("return []byte{")
			for j, value := range data {
				if j > 0 {
					literal.WriteString(", ")
				}
				fmt.Fprintf(&literal, "0x%02x", value)
			}
			literal.WriteString("}")
			lines = append(lines, literal.String())
			want := "func() []byte {\n  " + strings.Join(lines, "\n  ") + "\n}()"
			if got != want {
				t.Fatalf("case %d: changed rich byte formatting", i)
			}
		}
		fmt.Fprintf(&source, "{ got := %s; want, _ := base64.StdEncoding.DecodeString(%q); if !bytes.Equal(got, want) || (got == nil) != %t { panic(fmt.Sprintf(\"case %d did not round trip\")) } }\n", got, base64.StdEncoding.EncodeToString(data), data == nil, i)
	}
	nested := dd.Dump(struct{ Bytes []byte }{allBytes[:1]}, df.WithRichBytes(), dd.WithIndent(4))
	if !strings.Contains(nested, "\n        // 00000000") {
		t.Fatalf("missing nested indentation: %s", nested)
	}
	fmt.Fprintf(&source, "{ got := %s; if !bytes.Equal(got.Bytes, []byte{0}) { panic(\"nested bytes did not round trip\") } }\n}\n", nested)
	path := filepath.Join(t.TempDir(), "main.go")
	if err := ioutil.WriteFile(path, []byte(source.String()), 0600); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command(filepath.Join(runtime.GOROOT(), "bin", "go"), "run", path).CombinedOutput(); err != nil {
		t.Fatalf("generated program failed: %v\n%s", err, out)
	}
}

type richBytesFailure struct{ err error }

func (w richBytesFailure) Write(p []byte) (int, error) { return 0, w.err }

func TestRichBytesWriterError(t *testing.T) {
	sentinel := errors.New("output failed")
	for _, want := range []error{sentinel, io.ErrShortWrite} {
		writeErr := want
		if want == io.ErrShortWrite {
			writeErr = nil
		}
		if got := dd.DumpTo(richBytesFailure{writeErr}, make([]byte, 65537), df.WithRichBytes()); got != want {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}
