package data

import "reflect"

// isExported reports whether the field is exported.
func isExported(f reflect.StructField) bool {
	return f.PkgPath == ""
}

// valueOf returns a new Value initialized to the concrete value.
// returns obj if obj is reflect.Value.
func valueOf(obj interface{}) reflect.Value {
	if v, ok := obj.(reflect.Value); ok {
		return v
	}
	return reflect.ValueOf(obj)
}
