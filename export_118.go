package dd

import (
	"reflect"
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
