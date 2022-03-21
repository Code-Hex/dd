package dd

import (
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"
	"strings"
	"time"
)

// DumpFunc is a function to dump you specified custom format.
type DumpFunc[T any] func(T, Writer)

// WithDumpFunc is an option to add function for customize specified type dump string.
func WithDumpFunc[T any](f DumpFunc[T]) OptionFunc {
	var v T
	typ := reflect.TypeOf(v)
	return func(o *options) {
		o.convertibleTypes[typ] = dumpFunc(func(rv reflect.Value, w Writer) {
			f(rv.Interface().(T), w)
		})
	}
}

// WithJSONRawMessage is a wrapper of WithDumpFunc for json.RawMessage.
// Dumps a raw JSON string.
func WithJSONRawMessage() OptionFunc {
	return WithDumpFunc(
		func(v json.RawMessage, w Writer) {
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
func WithTime(format string) OptionFunc {
	return WithDumpFunc(
		func(v time.Time, w Writer) {
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
func WithBigInt() OptionFunc {
	return WithDumpFunc(
		func(v *big.Int, w Writer) {
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
func WithBigFloat() OptionFunc {
	return WithDumpFunc(
		func(v *big.Float, w Writer) {
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
