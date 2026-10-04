package dd

import (
	"io"
	"reflect"
	"strings"
)

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
	var out strings.Builder
	_ = DumpTo(&out, data, opts...)
	return out.String()
}

// DumpTo writes the dump to w without adding a trailing newline.
// It returns the first write error, including io.ErrShortWrite. Output may be partial.
// Slices and structs stream directly; map alignment and custom formatters may buffer.
// The caller owns w and is responsible for flushing or closing it.
func DumpTo(w io.Writer, data interface{}, opts ...OptionFunc) error {
	d := newDataDumper(&checkedWriter{Writer: w}, data, opts...)
	d.build()
	return d.err
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

// WithOmitEmptyFields omits struct fields whose values are zero according to
// reflect.Value.IsZero. Non-nil empty slices and maps, non-nil pointers, and
// interfaces containing typed nil values are retained. It applies recursively.
func WithOmitEmptyFields() OptionFunc {
	return func(o *options) {
		o.omitEmptyFields = true
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

// checkedWriter also catches short writes made while flushing a tabwriter.
type checkedWriter struct {
	io.Writer
	scratch []byte
}

func (w *checkedWriter) Write(p []byte) (int, error) {
	n, err := w.Writer.Write(p)
	if err == nil && n != len(p) {
		err = io.ErrShortWrite
	}
	return n, err
}
func (w *checkedWriter) WriteString(s string) (int, error) {
	if sw, ok := w.Writer.(io.StringWriter); ok {
		n, err := sw.WriteString(s)
		if err == nil && n != len(s) {
			err = io.ErrShortWrite
		}
		return n, err
	}
	// Reuse bounded storage instead of allocating []byte(s) for every token.
	if w.scratch == nil {
		w.scratch = make([]byte, 4096)
	}
	total := 0
	for len(s) > 0 {
		size := copy(w.scratch, s)
		n, err := w.Write(w.scratch[:size])
		total += n
		if err != nil {
			return total, err
		}
		s = s[size:]
	}
	return total, nil
}
