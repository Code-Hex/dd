package data

import (
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/Code-Hex/go-data-dumper/internal/sort"
)

type dumper struct {
	buf          *strings.Builder
	tw           *tabwriter.Writer
	value        reflect.Value
	depth        int
	exportedOnly bool
}

var _ interface {
	fmt.Stringer
} = (*dumper)(nil)

const indentSize = 2

func newDataDumper(obj interface{}, opts ...OptionFunc) *dumper {
	buf := new(strings.Builder)
	ret := &dumper{
		buf:          buf,
		tw:           tabwriter.NewWriter(buf, indentSize, 0, 1, ' ', 0),
		value:        valueOf(obj),
		depth:        0,
		exportedOnly: false,
	}
	// apply options
	for _, opt := range opts {
		opt(ret)
	}
	return ret
}

func (d *dumper) indent() string {
	return strings.Repeat("\t", d.depth)
}

func (d *dumper) String() string {
	d.tw.Flush()
	return d.buf.String()
}

func (d *dumper) build() *dumper {
	kind := d.value.Kind()
	switch kind {
	case reflect.Invalid:
		return d.writeRaw("nil")
	case reflect.Bool:
		return d.writeBool(d.value.Bool())
	case reflect.String:
		return d.writeString(d.value.String())
	case reflect.Array:
		return d.writeArray()
	case reflect.Slice:
		return d.writeSlice()
	case reflect.Map:
		return d.writeMap()
	case reflect.Chan:
		return d.writeChan()
	case reflect.Func:
		return d.printf("%s {...}", d.value.Type().String())
	case reflect.Struct:
		return d.writeStruct()
	case reflect.UnsafePointer:
		return d.printf("%s(%v)", d.value.Type().String(), d.value.Pointer())
	}
	if isNumber(kind) {
		return d.writeNumber()
	}
	return d.writeInterface()
}

func (d *dumper) clone(obj interface{}) *dumper {
	child := newDataDumper(obj)
	child.depth = d.depth
	child.exportedOnly = d.exportedOnly
	return child.build()
}

func (d *dumper) writeStruct() *dumper {
	numField := d.value.NumField()

	// records the i'th field
	fieldIdxs := make([]int, 0, numField)

	for i := 0; i < numField; i++ {
		field := d.value.Type().Field(i)
		if d.exportedOnly && !isExported(field) {
			continue
		}
		fieldIdxs = append(fieldIdxs, i)
	}
	if len(fieldIdxs) == 0 {
		return d.printf("%s{}", d.value.Type().String())
	}

	d.printf("%s{\n", d.value.Type().String())
	d.depth++
	for _, idx := range fieldIdxs {
		field := d.value.Type().Field(idx)
		fieldVal := d.value.Field(idx)
		dumper := d.clone(fieldVal)
		d.writeIndent()
		d.printf("%s:\t%s,\n", field.Name, dumper.String())
	}
	d.depth--
	return d.writeRaw("}")
}

func (d *dumper) writeChan() *dumper {
	d.writeRaw("(").writeChanType().writeRaw(")")

	if d.value.IsNil() {
		return d.writeRaw("(nil)")
	}
	return d.printf("(%v)", d.value.Pointer())
}

func (d *dumper) writeChanType() *dumper {
	switch d.value.Type().ChanDir() {
	case reflect.RecvDir:
		return d.printf("<-chan %s", d.value.Type().Elem().String())
	case reflect.SendDir:
		return d.printf("chan<- %s", d.value.Type().Elem().String())
	case reflect.BothDir:
		return d.writeRaw(d.value.Type().String())
	}
	panic("unreachable")
}

func (d *dumper) writeMap() *dumper {
	// We must check if it is nil before checking length.
	// because the length of nil map is 0.
	if d.value.IsNil() {
		return d.printf("(%s)(nil)", d.value.Type().String())
	}
	if d.value.Len() == 0 {
		return d.printf("%s{}", d.value.Type().String())
	}
	d.printf("%s{\n",
		d.value.Type().String(),
	)

	d.depth++
	for _, key := range sort.Keys(d.value.MapKeys()) {
		val := d.value.MapIndex(key)
		keyDumper := d.clone(key)
		valDumper := d.clone(val)
		d.writeIndent()
		d.printf("%s:\t%s,\n",
			keyDumper.String(),
			valDumper.String(),
		)
	}
	d.depth--
	return d.writeRaw("}")
}

func (d *dumper) writeSlice() *dumper {
	// We must check if it is nil before checking length.
	// because the length of nil slice is 0.
	if d.value.IsNil() {
		return d.printf("(%s)(nil)", d.value.Type().String())
	}
	return d.writeArray()
}

func (d *dumper) writeArray() *dumper {
	if d.value.Len() == 0 {
		return d.printf("%s{}", d.value.Type().String())
	}
	d.printf("%s{\n", d.value.Type().String())
	return d.writeList().writeRaw("}")
}

func (d *dumper) writeList() *dumper {
	d.depth++
	for i := 0; i < d.value.Len(); i++ {
		elem := d.value.Index(i)
		dumper := d.clone(elem)
		d.writeIndent()
		d.printf("%s,\n", dumper.String())
	}
	d.depth--
	return d
}

func (d *dumper) writeInterface() *dumper {
	elem := d.value.Elem()
	// immediate nil value which is like `var a = interface{}(nil)`
	// NOTE(codehex): maybe unnecessary?
	if elem.Kind() == reflect.Invalid {
		return d.writeRaw("nil")
	}
	if elem.IsValid() {
		return d.clone(elem)
	}
	return d.printf("(*%s)(nil)", elem.Type().String())
}

func (d *dumper) writeNumber() *dumper {
	switch d.value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return d.printf("%d", d.value.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return d.writeUnsignedInt(Decimal)
	case reflect.Float32, reflect.Float64:
		return d.printf("%f", d.value.Float())
	case reflect.Complex64:
		return d.printf("%v", complex64(d.value.Complex()))
	case reflect.Complex128:
		return d.printf("%v", d.value.Complex())
	}
	return d.printf("%#v", d.value.Interface())
}

const (
	Decimal = iota
	Binary
	Hex
)

func (d *dumper) writeUnsignedInt(typ int) *dumper {
	switch d.value.Kind() {
	case reflect.Uint8:
		switch typ {
		case Binary:
			return d.printf("0b%08b", d.value.Uint())
		case Hex:
			return d.printf("0b%02x", d.value.Uint())
		}
	case reflect.Uint16:
		switch typ {
		case Binary:
			return d.printf("0b%016b", d.value.Uint())
		case Hex:
			return d.printf("0b%04x", d.value.Uint())
		}
	case reflect.Uint32:
		switch typ {
		case Binary:
			return d.printf("0b%032b", d.value.Uint())
		case Hex:
			return d.printf("0b%08x", d.value.Uint())
		}
	case reflect.Uint64:
		switch typ {
		case Binary:
			return d.printf("0b%064b", d.value.Uint())
		case Hex:
			return d.printf("0b%016x", d.value.Uint())
		}
	}
	return d.writeRaw(strconv.FormatUint(d.value.Uint(), 10))
}

func (d *dumper) writeBool(b bool) *dumper {
	return d.writeRaw(strconv.FormatBool(b))
}

func (d *dumper) writeString(s string) *dumper {
	return d.writeRaw(strconv.Quote(s))
}

func (d *dumper) writeIndent() *dumper {
	if d.depth == 0 {
		return d
	}
	return d.writeRaw(d.indent())
}

// writeRaw appends the contents of s to p's buffer.
func (d *dumper) writeRaw(s string) *dumper {
	io.WriteString(d.tw, s)
	return d
}

func (d *dumper) printf(format string, a ...interface{}) *dumper {
	fmt.Fprintf(d.tw, format, a...)
	return d
}
