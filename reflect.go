package dd

import "reflect"

// valueOf returns a new Value initialized to the concrete value.
// returns obj if obj is reflect.Value.
func valueOf(obj interface{}) reflect.Value {
	if v, ok := obj.(reflect.Value); ok {
		return v
	}
	return reflect.ValueOf(obj)
}
