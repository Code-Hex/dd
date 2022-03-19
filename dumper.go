package dd

import (
	"bufio"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/Code-Hex/dd/internal/sort"
)

type dumpFunc func(reflect.Value, Writer)

type options struct {
	exportedOnly     bool
	indentSize       int
	uintFormat       UintFormat
	convertibleTypes map[reflect.Type]dumpFunc
}

func newDefaultOptions() *options {
	return &options{
		exportedOnly:     false,
		indentSize:       2,
		uintFormat:       DecimalUint,
		convertibleTypes: map[reflect.Type]dumpFunc{},
	}
}

type dumper[T any] struct {
	buf           *strings.Builder
	tw            *tabwriter.Writer
	value         reflect.Value
	v             T
	depth         int
	visitPointers map[uintptr]bool
	// options
	exportedOnly     bool
	uintFormat       UintFormat
	convertibleTypes map[reflect.Type]dumpFunc
}

var _ interface {
	fmt.Stringer
} = (*dumper[any])(nil)

func newDataDumper[T any](obj T, checkConcreteValue bool, optFuncs ...OptionFunc) *dumper[T] {
	buf := new(strings.Builder)
	opts := newDefaultOptions()
	// apply options
	for _, apply := range optFuncs {
		apply(opts)
	}
	return &dumper[T]{
		buf:              buf,
		tw:               tabwriter.NewWriter(buf, opts.indentSize, 0, 1, ' ', 0),
		value:            valueOf(obj, checkConcreteValue),
		v:                obj,
		depth:            0,
		visitPointers:    make(map[uintptr]bool),
		exportedOnly:     opts.exportedOnly,
		uintFormat:       opts.uintFormat,
		convertibleTypes: opts.convertibleTypes,
	}
}

func clone[T, U any](d *dumper[T], obj U) *dumper[U] {
	child := newDataDumper(obj, false)
	child.depth = d.depth
	child.visitPointers = d.visitPointers
	child.exportedOnly = d.exportedOnly
	child.uintFormat = d.uintFormat
	child.convertibleTypes = d.convertibleTypes
	return child.build()
}

func (d *dumper[T]) indent() string {
	return strings.Repeat("\t", d.depth)
}

func (d *dumper[T]) String() string {
	d.tw.Flush()
	return d.buf.String()
}

func (d *dumper[T]) build() *dumper[T] {
	kind := d.value.Kind()
	if kind == reflect.Invalid {
		return d.writeRaw("nil")
	}

	convertFunc, ok := d.convertibleTypes[d.value.Type()]
	if ok {
		convertFunc(d.value, &dumpWriter[T]{d})
		return d
	}
	switch kind {
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
		return d.printf("%s(uintptr(%v))", d.value.Type().String(), d.value.Pointer())
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

func (d *dumper[T]) writeFunc() *dumper[T] {
	if d.value.IsNil() {
		return d.printf("(%s)(nil)", d.value.Type().String())
	}

	if ret, ok := d.writeVisitedPointer(); ok {
		return ret
	}

	d.writeRaw(d.value.Type().String())
	d.writeRaw(" ")

	return d.writeBlock(func() {
		// function body
		d.writeIndentedRaw("// ...\n")
		typ := d.value.Type()
		numout := typ.NumOut()
		if numout == 0 {
			return
		}

		zeroValues := make([]string, 0, numout)
		for i := 0; i < numout; i++ {
			outTyp := typ.Out(i)
			zeroValues = append(
				zeroValues,
				clone(d, reflect.Zero(outTyp)).String(),
			)
		}
		d.indentedPrintf("return %s\n", strings.Join(zeroValues, ", "))
	})
}

func (d *dumper[T]) writePtr() *dumper[T] {
	if d.value.IsNil() {
		return d.printf("(%s)(nil)", d.value.Type())
	}
	if ret, ok := d.writeVisitedPointer(); ok {
		return ret
	}

	// dereference
	deref := d.value.Elem()
	kind := deref.Kind()
	if kind == reflect.Ptr {
		return d.writePointer()
	}
	if isPrimitive(kind) {
		return d.writePointer()
	}
	convertFunc, ok := d.convertibleTypes[deref.Type()]
	if ok {
		convertFunc(d.value, &dumpWriter[T]{d})
		return d
	}
	return d.printf("&%s", clone(d, deref).String())
}

func (d *dumper[T]) writeStruct() *dumper[T] {
	numField := d.value.NumField()

	// records the i'th field
	fieldIdxs := make([]int, 0, numField)

	for i := 0; i < numField; i++ {
		field := d.value.Type().Field(i)
		if d.exportedOnly && !field.IsExported() {
			continue
		}
		fieldIdxs = append(fieldIdxs, i)
	}
	if len(fieldIdxs) == 0 {
		return d.printf("%s{}", d.value.Type().String())
	}

	d.writeRaw(d.value.Type().String())
	return d.writeBlock(func() {
		for _, idx := range fieldIdxs {
			field := d.value.Type().Field(idx)
			fieldVal := d.value.Field(idx)
			dumper := clone(d, fieldVal)
			d.indentedPrintf("%s: %s,\n", field.Name, dumper.String())
		}
	})
}

// writeChan writes channel info. format will be like `(chan int)(nil)`
func (d *dumper[T]) writeChan() *dumper[T] {
	if d.value.IsNil() {
		return d.printf("(%s)(nil)", d.value.Type().String())
	}
	return d.writePointer()
}

func (d *dumper[T]) writeMap() *dumper[T] {
	// We must check if it is nil before checking length.
	// because the length of nil map is 0.
	if d.value.IsNil() {
		return d.printf("(%s)(nil)", d.value.Type().String())
	}
	if d.value.Len() == 0 {
		return d.printf("%s{}", d.value.Type().String())
	}

	if ret, ok := d.writeVisitedPointer(); ok {
		return ret
	}

	d.writeRaw(d.value.Type().String())

	return d.writeBlock(func() {
		for _, key := range sort.Keys(d.value.MapKeys()) {
			val := d.value.MapIndex(key)
			keyDumper := clone(d, key)
			valDumper := clone(d, val)
			d.indentedPrintf("%s:\t%s,\n",
				keyDumper.String(),
				valDumper.String(),
			)
		}
	})
}

func (d *dumper[T]) writeSlice() *dumper[T] {
	// We must check if it is nil before checking length.
	// because the length of nil slice is 0.
	if d.value.IsNil() {
		return d.printf("(%s)(nil)", d.value.Type().String())
	}

	if ret, ok := d.writeVisitedPointer(); ok {
		return ret
	}

	return d.writeArray()
}

func (d *dumper[T]) writeArray() *dumper[T] {
	if d.value.Len() == 0 {
		return d.printf("%s{}", d.value.Type().String())
	}
	d.writeRaw(d.value.Type().String())
	return d.writeList()
}

func (d *dumper[T]) writeList() *dumper[T] {
	return d.writeBlock(func() {
		for i := 0; i < d.value.Len(); i++ {
			elem := d.value.Index(i)
			dumper := clone(d, elem)
			d.indentedPrintf("%s,\n", dumper.String())
		}
	})
}

func (d *dumper[T]) writeInterface() *dumper[T] {
	elem := d.value.Elem()
	if elem.IsValid() {
		return d.writeRaw(clone(d, elem).String())
	}
	return d.writeRaw("nil")
}

func (d *dumper[T]) writeNumber() *dumper[T] {
	switch d.value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return d.printf("%d", d.value.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return d.writeUnsignedInt()
	case reflect.Float32, reflect.Float64:
		return d.printf("%f", d.value.Float())
	case reflect.Complex64:
		return d.printf("%v", complex64(d.value.Complex()))
	case reflect.Complex128:
		return d.printf("%v", d.value.Complex())
	}
	panic(fmt.Errorf("unreachable type: %s", d.value.Type()))
}

func (d *dumper[T]) writeUnsignedInt() *dumper[T] {
	switch d.value.Kind() {
	case reflect.Uint8:
		switch d.uintFormat {
		case BinaryUint:
			return d.printf("0b%08b", d.value.Uint())
		case HexUint:
			return d.printf("0x%02x", d.value.Uint())
		}
	case reflect.Uint16:
		switch d.uintFormat {
		case BinaryUint:
			return d.printf("0b%016b", d.value.Uint())
		case HexUint:
			return d.printf("0x%04x", d.value.Uint())
		}
	case reflect.Uint32:
		switch d.uintFormat {
		case BinaryUint:
			return d.printf("0b%032b", d.value.Uint())
		case HexUint:
			return d.printf("0x%08x", d.value.Uint())
		}
	case reflect.Uint64:
		switch d.uintFormat {
		case BinaryUint:
			return d.printf("0b%064b", d.value.Uint())
		case HexUint:
			return d.printf("0x%016x", d.value.Uint())
		}
	}
	return d.writeRaw(strconv.FormatUint(d.value.Uint(), 10))
}

func (d *dumper[T]) writePointer() *dumper[T] {
	return d.printf(
		"(%s)(unsafe.Pointer(uintptr(0x%x)))",
		d.value.Type().String(),
		d.value.Pointer(),
	)
}

func (d *dumper[T]) writeVisitedPointer() (*dumper[T], bool) {
	pointer := d.value.Pointer()
	if d.visitPointers[pointer] {
		return d.writePointer(), true
	}
	d.visitPointers[pointer] = true
	return d, false
}

func (d *dumper[T]) writeBlock(f func()) *dumper[T] {
	d.writeRaw("{\n")
	d.depth++
	f()
	d.depth--
	return d.writeIndentedRaw("}")
}

func (d *dumper[T]) writeBool(b bool) *dumper[T] {
	return d.writeRaw(strconv.FormatBool(b))
}

func (d *dumper[T]) writeString(s string) *dumper[T] {
	return d.writeRaw(strconv.Quote(s))
}

func (d *dumper[T]) writeIndent() *dumper[T] {
	if d.depth == 0 {
		return d
	}
	return d.writeRaw(d.indent())
}

func (d *dumper[T]) writeIndentedRaw(s string) *dumper[T] {
	d.writeIndent()
	return d.writeRaw(s)
}

func (d *dumper[T]) indentedPrintf(format string, a ...interface{}) *dumper[T] {
	d.writeIndent()
	return d.printf(format, a...)
}

// writeRaw appends the contents of s to p's buffer.
func (d *dumper[T]) writeRaw(s string) *dumper[T] {
	io.WriteString(d.tw, s)
	return d
}

func (d *dumper[T]) printf(format string, a ...interface{}) *dumper[T] {
	fmt.Fprintf(d.tw, format, a...)
	return d
}

type dumpWriter[T any] struct{ *dumper[T] }

var _ Writer = (*dumpWriter[any])(nil)

func (d *dumpWriter[T]) Write(s string) { d.dumper.writeRaw(s) }
func (d *dumpWriter[T]) WriteBlock(s string) {
	d.dumper.writeRaw("{\n")
	d.dumper.depth++
	scanner := bufio.NewScanner(strings.NewReader(s))
	for scanner.Scan() {
		d.dumper.writeIndentedRaw(scanner.Text() + "\n")
	}
	d.dumper.depth--
	d.dumper.writeIndentedRaw("}")
}
