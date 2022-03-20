package dd

import "reflect"

var zeroPrimitives = map[reflect.Type]string{
	reflect.TypeOf(bool(false)):          "false",
	reflect.TypeOf(uint8(0x0)):           "0",
	reflect.TypeOf(uint16(0x0)):          "0",
	reflect.TypeOf(uint32(0x0)):          "0",
	reflect.TypeOf(uint64(0x0)):          "0",
	reflect.TypeOf(int8(0)):              "0",
	reflect.TypeOf(int16(0)):             "0",
	reflect.TypeOf(int32(0)):             "0",
	reflect.TypeOf(int64(0)):             "0",
	reflect.TypeOf(float32(0)):           "0.000000",
	reflect.TypeOf(float64(0)):           "0.000000",
	reflect.TypeOf(complex64((0 + 0i))):  "(0+0i)",
	reflect.TypeOf(complex128((0 + 0i))): "(0+0i)",
	reflect.TypeOf(string("")):           "\"\"",
	reflect.TypeOf(int(0)):               "0",
	reflect.TypeOf(uint(0x0)):            "0",
	reflect.TypeOf(uintptr(0x0)):         "0",
}
