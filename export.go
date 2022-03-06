package dd

import (
	"fmt"
	"math/big"
	"reflect"
	"strings"
	"time"
)

// Writer is a writer for dump string.
type Writer interface {
	Write(s string)
	WriteBlock(s string)
}

// DumpFunc is a function to dump you specified custom format.
type DumpFunc func(reflect.Value, Writer)

type OptionFunc func(*options)

func WithExportedOnly() OptionFunc {
	return func(o *options) {
		o.exportedOnly = true
	}
}

func WithIndent(indent int) OptionFunc {
	return func(o *options) {
		o.indentSize = indent
	}
}

// WithTime is a wrapper of WithDumpFunc for time.Time.
// Dumps the numeric values instead of displaying the struct contents.
func WithTime(format string) OptionFunc {
	return WithDumpFunc(
		reflect.TypeOf(time.Time{}),
		func(rv reflect.Value, w Writer) {
			tmp := rv.Interface().(time.Time)
			w.Write("func() time.Time ")
			w.WriteBlock(
				strings.Join(
					[]string{
						fmt.Sprintf(
							"tmp, _ := time.Parse(%q, %q)",
							format, tmp.Format(format),
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
		reflect.TypeOf(big.Int{}),
		func(rv reflect.Value, w Writer) {
			tmp := rv.Interface().(*big.Int)
			w.Write("func() *big.Int ")
			w.WriteBlock(
				strings.Join(
					[]string{
						"tmp := new(big.Int)",
						fmt.Sprintf(
							"tmp.SetString(%q)",
							tmp.String(),
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
		reflect.TypeOf(big.Float{}),
		func(rv reflect.Value, w Writer) {
			tmp := rv.Interface().(*big.Float)
			w.Write("func() *big.Float ")
			w.WriteBlock(
				strings.Join(
					[]string{
						"tmp := new(big.Float)",
						fmt.Sprintf(
							"tmp.SetString(%q)",
							tmp.String(),
						),
						"return tmp",
					},
					"\n",
				),
			)
		},
	)
}

// WithDumpFunc is an option to add function for customize specified type dump string.
func WithDumpFunc(target reflect.Type, f DumpFunc) OptionFunc {
	return func(o *options) {
		o.convertibleTypes[target] = f
	}
}

func Dump(v interface{}, opts ...OptionFunc) string {
	return newDataDumper(v, opts...).build().String()
}
