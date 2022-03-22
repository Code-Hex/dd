package df

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/Code-Hex/dd"
)

// WithBytes is a wrapper of WithDumpFunc for []byte and []uint8.
// The format of the dump matches the output of `hexdump -C` on the command line.
func WithRichBytes() dd.OptionFunc {
	return dd.WithDumpFunc(
		func(v []byte, w dd.Writer) {
			dumpLines := strings.Split(hex.Dump(v), "\n")
			for i := 0; i < len(dumpLines); i++ {
				dumpLines[i] = "// " + dumpLines[i]
			}
			var buf strings.Builder
			buf.WriteString("return []byte{")
			for i, b := range v {
				fmt.Fprintf(&buf, "0x%02x", b)
				if i != len(v)-1 {
					buf.WriteString(", ")
				}
			}
			buf.WriteString("}")
			dumpLines = append(dumpLines, buf.String())
			w.Write("func() []byte ")
			w.WriteBlock(strings.Join(dumpLines, "\n"))
			w.Write("()")
		},
	)
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
