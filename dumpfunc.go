//go:build !go1.18
// +build !go1.18

package dd

import "reflect"

type any = interface{}

// DumpFunc is a function to dump you specified custom format.
type DumpFunc func(interface{}, Writer)

var typeWriter = reflect.TypeOf((*Writer)(nil)).Elem()

// WithDumpFunc is an option to add function for customize specified type dump string.
// want function f like "func(string, Writer)"
func WithDumpFunc(f interface{}) OptionFunc {
	frv := reflect.ValueOf(f)
	typ := frv.Type()
	if typ.Kind() != reflect.Func {
		panic("f must be function")
	}
	numIn := typ.NumIn()
	if numIn != 2 {
		panic("f must be the number of parameter is 2")
	}
	p0 := typ.In(0)
	p1 := typ.In(1)
	if !p1.Implements(typeWriter) {
		panic("the second parameter must be implemented interface Writer")
	}
	return func(o *options) {
		o.convertibleTypes[p0] = dumpFunc(func(rv reflect.Value, w Writer) {
			frv.Call([]reflect.Value{rv, reflect.ValueOf(w)})
		})
	}
}
