package dd

import (
	"fmt"
	"math/big"
	"reflect"
	"strings"
	"time"
)

type UintFormat int

const (
	// DecimalUint is mode to display uint as decimal format
	DecimalUint UintFormat = iota
	// BinaryUint is mode to display uint as binary format
	// The format be like 0b00000000
	BinaryUint
	// HexUint is mode to display uint as hex format
	// The format be like 0x00
	HexUint
)

// Dump dumps specified data.
func Dump(data interface{}, opts ...OptionFunc) string {
	return newDataDumper(data, opts...).build().String()
}

// Writer is a writer for dump string.
type Writer interface {
	Write(s string)
	WriteBlock(s string)
}

// DumpFunc is a function to dump you specified custom format.
type DumpFunc func(reflect.Value, Writer)

// OptionFunc is a function for making options.
type OptionFunc func(*options)

// WithExportedOnly enables to display only exported struct field.
// ignores unexported field.
func WithExportedOnly() OptionFunc {
	return func(o *options) {
		o.exportedOnly = true
	}
}

// WithIndent adjust indent nested in any blocks.
// default is 2 spaces.
func WithIndent(indent int) OptionFunc {
	return func(o *options) {
		o.indentSize = indent
	}
}

// WithUintFormat specify mode to display uint format.
// default is DecimalUint.
func WithUintFormat(mode UintFormat) OptionFunc {
	return func(o *options) {
		o.uintFormat = mode
	}
}

// WithBytes is a wrapper of WithDumpFunc for []byte.
// Dumps the byte slice values instead of displaying uint8 slice.
func WithBytes(mode UintFormat) OptionFunc {
	return WithDumpFunc(
		reflect.TypeOf([]byte{}),
		func(rv reflect.Value, w Writer) {
			tmp := rv.Interface().([]byte)
			w.Write("[]byte")
			var buf strings.Builder
			for _, b := range tmp {
				switch mode {
				case BinaryUint:
					fmt.Fprintf(&buf, "0b%08b,\n", b)
				default:
					fmt.Fprintf(&buf, "0x%02x,\n", b)
				}
			}
			w.WriteBlock(buf.String())
		},
	)
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
