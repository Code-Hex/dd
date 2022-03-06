package dd

import "reflect"

func isPrimitive(kind reflect.Kind) bool {
	return kind == reflect.String || kind == reflect.Bool || isNumber(kind)
}

func isNumber(kind reflect.Kind) bool {
	return isInt(kind) || isUint(kind) || isFloat(kind) || isComplex(kind)
}

func isInt(kind reflect.Kind) bool {
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	}
	return false
}

func isUint(kind reflect.Kind) bool {
	switch kind {
	case reflect.Uintptr, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	}
	return false
}

func isFloat(kind reflect.Kind) bool {
	return kind == reflect.Float32 || kind == reflect.Float64
}

func isComplex(kind reflect.Kind) bool {
	return kind == reflect.Complex64 || kind == reflect.Complex128
}
