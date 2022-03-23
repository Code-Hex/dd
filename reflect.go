package dd

import (
	"reflect"
	"unsafe"
)

// valueOf returns a new Value initialized to the concrete value.
// returns obj if obj is reflect.Value.
func valueOf(obj interface{}, checkConcrete bool) reflect.Value {
	if v, ok := obj.(reflect.Value); ok && !checkConcrete {
		return v
	}
	return reflect.ValueOf(obj)
}

func isExported(f reflect.StructField) bool {
	return f.PkgPath == ""
}

// https://stackoverflow.com/a/43918797
func getUnexportedField(f reflect.Value) reflect.Value {
	return reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Elem()
}
