package df

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/Code-Hex/dd"
)

// WithRichBytes is a wrapper of WithDumpFunc for []byte and []uint8.
// The format of the dump matches the output of `hexdump -C` on the command line.
// Formatting retains a complete output block in memory.
func WithRichBytes() dd.OptionFunc {
	return dd.WithDumpFunc(
		func(v []byte, w dd.Writer) {
			if v == nil {
				w.Write("[]byte(nil)")
				return
			}
			var buf strings.Builder
			// Reserve the block once: at most 82 bytes per comment line,
			// six per byte literal, and the return statement punctuation.
			if len(v) <= (int(^uint(0)>>1)-98)/12 {
				buf.Grow((len(v)/16+1)*82 + len(v)*6 + 16)
			}
			dumper := hex.Dumper(&commentWriter{out: &buf, lineStart: true})
			_, _ = dumper.Write(v)
			_ = dumper.Close()
			buf.WriteString("\nreturn []byte{")
			const digits = "0123456789abcdef"
			for i, b := range v {
				if i > 0 {
					buf.WriteString(", ")
				}
				buf.WriteString("0x")
				buf.WriteByte(digits[b>>4])
				buf.WriteByte(digits[b&15])
			}
			buf.WriteString("}")
			w.Write("func() []byte ")
			w.WriteBlock(buf.String())
			w.Write("()")
		},
	)
}

// commentWriter prefixes hex.Dumper's lines without retaining a second copy.
type commentWriter struct {
	out       *strings.Builder
	lineStart bool
}

func (w *commentWriter) Write(p []byte) (int, error) {
	n := len(p)
	for len(p) > 0 {
		if w.lineStart {
			w.out.WriteString("// ")
		}
		i := bytes.IndexByte(p, '\n')
		w.lineStart = i >= 0
		if i < 0 {
			i = len(p) - 1
		}
		w.out.Write(p[:i+1])
		p = p[i+1:]
	}
	return n, nil
}

// WithJSONRawMessage is a wrapper of WithDumpFunc for json.RawMessage.
// Dumps a raw JSON string.
func WithJSONRawMessage() dd.OptionFunc {
	return dd.WithDumpFunc(
		func(v json.RawMessage, w dd.Writer) {
			w.Write("json.RawMessage(")
			w.Write("`")
			w.Write(string(v))
			w.Write("`")
			w.Write(")")
		},
	)
}

// WithTime is a wrapper of WithDumpFunc for time.Time.
// Dumps the numeric values instead of displaying the struct contents.
func WithTime(format string) dd.OptionFunc {
	return dd.WithDumpFunc(
		func(v time.Time, w dd.Writer) {
			w.Write("func() time.Time ")
			w.WriteBlock(
				strings.Join(
					[]string{
						fmt.Sprintf(
							"tmp, _ := time.Parse(%q, %q)",
							format, v.Format(format),
						),
						"return tmp",
					},
					"\n",
				),
			)
			w.Write("()")
		},
	)
}

// WithBigInt is a wrapper of WithDumpFunc for big.Int.
// Dumps the numeric values instead of displaying the struct contents.
func WithBigInt() dd.OptionFunc {
	return dd.WithDumpFunc(
		func(v *big.Int, w dd.Writer) {
			w.Write("func() *big.Int ")
			w.WriteBlock(
				strings.Join(
					[]string{
						"tmp := new(big.Int)",
						fmt.Sprintf(
							"tmp.SetString(%q)",
							v.String(),
						),
						"return tmp",
					},
					"\n",
				),
			)
			w.Write("()")
		},
	)
}

// WithBigFloat is a wrapper of WithDumpFunc for big.Float.
// Dumps the numeric values instead of displaying the struct contents.
func WithBigFloat() dd.OptionFunc {
	return dd.WithDumpFunc(
		func(v *big.Float, w dd.Writer) {
			w.Write("func() *big.Float ")
			w.WriteBlock(
				strings.Join(
					[]string{
						"tmp := new(big.Float)",
						fmt.Sprintf(
							"tmp.SetString(%q)",
							v.String(),
						),
						"return tmp",
					},
					"\n",
				),
			)
			w.Write("()")
		},
	)
}
