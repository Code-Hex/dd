package dd

import "reflect"

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
