package dd

import "reflect"

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

// WithListBreakLineSize is an option to specify the number of elements to break lines
// when dumped a listing (slice, array) of a given type.
// The number must be more than 1 otherwise treats as 1.
func WithListBreakLineSize(typ interface{}, size int) OptionFunc {
	return func(o *options) {
		tmp := reflect.TypeOf(typ)
		o.listGroupingSize[tmp] = size
	}
}
