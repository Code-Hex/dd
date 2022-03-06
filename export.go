package dd

import (
	"math/big"
	"reflect"
	"strconv"
	"time"
)

// DumpFunc is a function to dump you specified custom format.
type DumpFunc func(reflect.Value) string

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
		func(rv reflect.Value) string {
			tmp := rv.Interface().(time.Time)
			return strconv.Quote(tmp.Format(format))
		},
	)
}

// WithBigInt is a wrapper of WithDumpFunc for big.Int.
// Dumps the numeric values instead of displaying the struct contents.
func WithBigInt() OptionFunc {
	return WithDumpFunc(
		reflect.TypeOf(big.Int{}),
		func(rv reflect.Value) string {
			tmp := rv.Interface().(big.Int)
			return tmp.String()
		},
	)
}

// WithBigFloat is a wrapper of WithDumpFunc for big.Float.
// Dumps the numeric values instead of displaying the struct contents.
func WithBigFloat() OptionFunc {
	return WithDumpFunc(
		reflect.TypeOf(big.Float{}),
		func(rv reflect.Value) string {
			tmp := rv.Interface().(big.Float)
			return tmp.String()
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
