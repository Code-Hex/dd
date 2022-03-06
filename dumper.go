package dd

import (
	"bufio"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/Code-Hex/go-data-dumper/internal/sort"
)

type options struct {
	exportedOnly     bool
	indentSize       int
	convertibleTypes map[reflect.Type]DumpFunc
}

func newDefaultOptions() *options {
	return &options{
		exportedOnly:     false,
		indentSize:       2,
		convertibleTypes: map[reflect.Type]DumpFunc{},
	}
}

type dumper struct {
	buf   *strings.Builder
	tw    *tabwriter.Writer
	value reflect.Value
	depth int
	// options
	exportedOnly     bool
	convertibleTypes map[reflect.Type]DumpFunc
}

var _ interface {
	fmt.Stringer
} = (*dumper)(nil)

func newDataDumper(obj interface{}, optFuncs ...OptionFunc) *dumper {
	buf := new(strings.Builder)
	opts := newDefaultOptions()
	// apply options
	for _, apply := range optFuncs {
		apply(opts)
	}
	return &dumper{
		buf:              buf,
		tw:               tabwriter.NewWriter(buf, opts.indentSize, 0, 1, ' ', 0),
		value:            valueOf(obj),
		depth:            0,
		exportedOnly:     opts.exportedOnly,
		convertibleTypes: opts.convertibleTypes,
	}
}

func (d *dumper) clone(obj interface{}) *dumper {
	child := newDataDumper(obj)
	child.depth = d.depth
	child.exportedOnly = d.exportedOnly
	child.convertibleTypes = d.convertibleTypes
	return child.build()
}

func (d *dumper) indent() string {
	return strings.Repeat("\t", d.depth)
}

func (d *dumper) String() string {
	d.tw.Flush()
	return d.buf.String()
}

func (d *dumper) build() *dumper {
	for typ, convertFunc := range d.convertibleTypes {
		if d.value.Type().ConvertibleTo(typ) {
			convertFunc(d.value, &dumpWriter{d})
			return d
		}
	}
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
		return d.writeFunc()
	case reflect.Struct:
		return d.writeStruct()
	case reflect.Interface:
		return d.writeInterface()
	case reflect.UnsafePointer:
		return d.printf("%s(%v)", d.value.Type().String(), d.value.Pointer())
	case reflect.Ptr:
		return d.writePtr()
	}
	if isNumber(kind) {
		return d.writeNumber()
	}
	if d.value.CanInterface() {
		return d.printf("%v", d.value.Interface())
	}
	return d.writeRaw(d.value.String())
}

func (d *dumper) writeFunc() *dumper {
	if d.value.IsNil() {
		return d.writeRaw("nil")
	}
	d.printf("%s {\n", d.value.Type().String())

	// function body
	d.depth++
	d.writeIndentedRaw("// ...\n")
	defer func() {
		d.depth--
		d.writeRaw("}")
	}()

	typ := d.value.Type()
	numout := typ.NumOut()
	if numout == 0 {
		return d.writeIndentedRaw("return\n")
	}

	zeroValues := make([]string, 0, numout)
	for i := 0; i < numout; i++ {
		outTyp := typ.Out(i)
		zeroValues = append(
			zeroValues,
			d.clone(reflect.Zero(outTyp)).String(),
		)
	}
	return d.indentedPrintf("return %s\n", strings.Join(zeroValues, ", "))
}

func (d *dumper) writePtr() *dumper {
	if d.value.IsNil() {
		return d.writeRaw("nil")
	}
	// dereference
	deref := d.value.Elem()
	kind := deref.Kind()
	if kind == reflect.Ptr {
		return d.printf(
			"(%s)(unsafe.Pointer(uintptr(0x%x)))",
			d.value.Type().String(),
			d.value.Pointer(),
		)
	}
	if isPrimitive(kind) {
		return d.printf(
			"(%s)(unsafe.Pointer(uintptr(0x%x)))",
			d.value.Type().String(),
			d.value.Pointer(),
		)
	}
	for typ, convertFunc := range d.convertibleTypes {
		if deref.Type().ConvertibleTo(typ) {
			convertFunc(d.value, &dumpWriter{d})
			return d
		}
	}
	return d.printf("&%s", d.clone(deref).String())
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
		d.indentedPrintf("%s:\t%s,\n", field.Name, dumper.String())
	}
	d.depth--
	return d.writeRaw("}")
}

// writeChan writes channel info. format will be like `(chan int)(nil)`
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
	return d.writeList().writeIndentedRaw("}")
}

func (d *dumper) writeList() *dumper {
	d.depth++
	for i := 0; i < d.value.Len(); i++ {
		elem := d.value.Index(i)
		dumper := d.clone(elem)
		d.indentedPrintf("%s,\n", dumper.String())
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

func (d *dumper) writeIndentedRaw(s string) *dumper {
	d.writeIndent()
	return d.writeRaw(s)
}

func (d *dumper) indentedPrintf(format string, a ...interface{}) *dumper {
	d.writeIndent()
	return d.printf(format, a...)
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

type dumpWriter struct{ *dumper }

var _ Writer = (*dumpWriter)(nil)

func (d *dumpWriter) Write(s string) { d.dumper.writeRaw(s) }
func (d *dumpWriter) WriteBlock(s string) {
	d.dumper.writeRaw("{\n")
	d.dumper.depth++
	scanner := bufio.NewScanner(strings.NewReader(s))
	for scanner.Scan() {
		d.dumper.writeIndentedRaw(scanner.Text() + "\n")
	}
	d.dumper.depth--
	d.dumper.writeIndentedRaw("}")
}
