package df

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/Code-Hex/dd"
)

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
		},
	)
}
