package dd

import (
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"text/tabwriter"
	"unicode/utf8"

	"github.com/Code-Hex/dd/internal/sort"
)

type dumpFunc func(reflect.Value, Writer)

type options struct {
	exportedOnly     bool
	omitEmptyFields  bool
	indentSize       int
	uintFormat       UintFormat
	convertibleTypes map[reflect.Type]dumpFunc
	listGroupingSize map[reflect.Type]int
}

func newDefaultOptions() *options {
	return &options{
		exportedOnly:     false,
		indentSize:       2,
		uintFormat:       DecimalUint,
		convertibleTypes: map[reflect.Type]dumpFunc{},
		listGroupingSize: map[reflect.Type]int{},
	}
}

// A slice view or a differently typed pointer can share an address without a cycle.
type visit struct {
	typ     reflect.Type
	pointer uintptr
	length  int
}

type dumper struct {
	scratch          []byte
	w                io.Writer
	err              error
	value            reflect.Value
	depth            int
	indentSize       int
	visitPointers    map[visit]bool
	exportedOnly     bool
	omitEmptyFields  bool
	uintFormat       UintFormat
	convertibleTypes map[reflect.Type]dumpFunc
	listGroupingSize map[reflect.Type]int
}

func newDataDumper(w io.Writer, obj interface{}, optFuncs ...OptionFunc) *dumper {
	opts := newDefaultOptions()
	for _, apply := range optFuncs {
		apply(opts)
	}
	if opts.indentSize < 0 {
		panic("negative indent")
	}
	if opts.indentSize == 0 {
		opts.indentSize = 1
	}
	return &dumper{
		w: w, value: valueOf(obj, true), indentSize: opts.indentSize,
		visitPointers: make(map[visit]bool),
		exportedOnly:  opts.exportedOnly, omitEmptyFields: opts.omitEmptyFields,
		uintFormat: opts.uintFormat, convertibleTypes: opts.convertibleTypes,
		listGroupingSize: opts.listGroupingSize,
	}
}

func (d *dumper) writeValue(value reflect.Value) {
	if d.err != nil {
		return
	}
	previous := d.value
	d.value = value
	d.build()
	d.value = previous
}

// Keep tab expansion local to maps and user-provided formatting.
func (d *dumper) withTabs(f func()) {
	if d.err != nil {
		return
	}
	previous := d.w
	tw := tabwriter.NewWriter(previous, d.indentSize, 0, 1, ' ', 0)
	d.w = tw
	f()
	if d.err == nil {
		d.err = tw.Flush()
	}
	d.w = previous
}

func (d *dumper) build() *dumper {
	if d.err != nil {
		return d
	}
	kind := d.value.Kind()
	if kind == reflect.Invalid {
		d.writeRaw("nil")
		return d
	}

	convertFunc, ok := d.convertibleTypes[d.value.Type()]
	if ok {
		d.withTabs(func() { convertFunc(d.value, &dumpWriter{d}) })
		return d
	}
	switch kind {
	case reflect.Bool:
		d.writeBool(d.value.Bool())
		return d
	case reflect.String:
		d.writeString(d.value.String())
		return d
	case reflect.Array:
		d.writeArray()
		return d
	case reflect.Slice:
		d.writeSlice()
		return d
	case reflect.Map:
		d.writeMap()
		return d
	case reflect.Chan:
		d.writeChan()
		return d
	case reflect.Func:
		d.writeFunc()
		return d
	case reflect.Struct:
		d.writeStruct()
		return d
	case reflect.Interface:
		d.writeInterface()
		return d
	case reflect.UnsafePointer:
		d.printf("%s(uintptr(%v))", d.value.Type().String(), d.value.Pointer())
		return d
	case reflect.Ptr:
		d.writePtr()
		return d
	}
	if isNumber(kind) {
		d.writeNumber()
		return d
	}
	// NOTE(codehex): perhaps this block is unnecessary
	if d.value.CanInterface() {
		d.printf("%v", d.value.Interface())
		return d
	}
	d.writeRaw(d.value.String())
	return d
}

func (d *dumper) writeFunc() {
	if d.value.IsNil() {
		d.printf("(%s)(nil)", d.value.Type().String())
		return
	}

	cleanup, ok := d.writeVisitedPointer()
	if ok {
		return
	}
	defer cleanup()

	typ := d.value.Type()
	funcTyp := typ.String()
	d.writeRaw(funcTyp)
	// check anonymous function or not.
	// e.g. context.CancelFunc => context.CancelFunc(func() {})
	isNotAnonymous := !strings.HasPrefix(funcTyp, "func(")
	if isNotAnonymous {
		d.writeRaw("(func(")
		for i := 0; i < typ.NumIn(); i++ {
			if i > 0 {
				d.writeRaw(", ")
			}
			if typ.IsVariadic() && i == typ.NumIn()-1 {
				d.writeRaw("..." + typ.In(i).Elem().String())
			} else {
				d.writeRaw(typ.In(i).String())
			}
		}
		d.writeRaw(")")
		switch typ.NumOut() {
		case 0:
		case 1:
			d.writeRaw(" " + typ.Out(0).String())
		default:
			d.writeRaw(" (")
			for i := 0; i < typ.NumOut(); i++ {
				if i > 0 {
					d.writeRaw(", ")
				}
				d.writeRaw(typ.Out(i).String())
			}
			d.writeRaw(")")
		}
	}
	d.writeRaw(" ")

	d.writeBlock(func() {
		// function body
		d.writeIndentedRaw("// ...\n")
		typ := d.value.Type()
		numout := typ.NumOut()
		if numout == 0 {
			return
		}

		d.writeIndentedRaw("return ")
		for i := 0; i < numout && d.err == nil; i++ {
			if i > 0 {
				d.writeRaw(", ")
			}
			d.writeZeroValue(typ.Out(i))
		}
		d.writeRaw("\n")
	})
	if isNotAnonymous {
		d.writeRaw(")")
	}
}

//go:generate go run cmd/zero/main.go

func (d *dumper) writeZeroValue(rt reflect.Type) {
	// Only default primitive output is independent of options and nesting depth.
	if len(d.convertibleTypes) == 0 && d.uintFormat == DecimalUint {
		if zero, ok := zeroPrimitives[rt]; ok {
			d.writeRaw(zero)
			return
		}
	}
	d.writeValue(reflect.Zero(rt))
}

func (d *dumper) writePtr() {
	if d.value.IsNil() {
		d.printf("(%s)(nil)", d.value.Type())
		return
	}
	cleanup, ok := d.writeVisitedPointer()
	if ok {
		return
	}
	defer cleanup()

	// dereference
	deref := d.value.Elem()
	kind := deref.Kind()
	if kind == reflect.Ptr {
		d.writePointer()
		return
	}
	if isPrimitive(kind) {
		d.writePointer()
		return
	}
	d.writeRaw("&")
	d.writeValue(deref)
}

func (d *dumper) writeStruct() {
	numField := d.value.NumField()

	// records the i'th field
	fieldIdxs := make([]int, 0, numField)

	for i := 0; i < numField; i++ {
		field := d.value.Type().Field(i)
		if d.exportedOnly && !isExported(field) {
			continue
		}
		if d.omitEmptyFields && d.value.Field(i).IsZero() {
			continue
		}
		fieldIdxs = append(fieldIdxs, i)
	}
	if len(fieldIdxs) == 0 {
		d.printf("%s{}", d.value.Type().String())
		return
	}

	d.writeRaw(d.value.Type().String())
	d.writeBlock(func() {
		for _, idx := range fieldIdxs {
			field := d.value.Type().Field(idx)
			fieldVal := d.value.Field(idx)
			if !isExported(field) && fieldVal.CanAddr() {
				fieldVal = getUnexportedField(fieldVal)
			}
			if d.err != nil {
				return
			}
			d.writeIndent()
			d.writeRaw(field.Name)
			d.writeRaw(": ")
			d.writeValue(fieldVal)
			d.writeRaw(",\n")
		}
	})
}

// writeChan writes channel info. format will be like `(chan int)(nil)`
func (d *dumper) writeChan() {
	if d.value.IsNil() {
		d.printf("(%s)(nil)", d.value.Type().String())
		return
	}
	d.writePointer()
}

func (d *dumper) writeMap() {
	// We must check if it is nil before checking length.
	// because the length of nil map is 0.
	if d.value.IsNil() {
		d.printf("(%s)(nil)", d.value.Type().String())
		return
	}
	if d.value.Len() == 0 {
		d.printf("%s{}", d.value.Type().String())
		return
	}

	cleanup, ok := d.writeVisitedPointer()
	if ok {
		return
	}
	defer cleanup()

	d.writeRaw(d.value.Type().String())

	d.withTabs(func() {
		d.writeBlock(func() {
			keys, values := sort.MapEntries(d.value)
			for i, key := range keys {
				if d.err != nil {
					return
				}
				val := values[i]
				// Preserve tabwriter's indentation columns for multiline keys.
				d.writeRaw(strings.Repeat("\t", d.depth))
				d.writeValue(key)
				d.writeRaw(":\t")
				d.writeValue(val)
				d.writeRaw(",\n")
			}
		})
	})
}

func (d *dumper) writeSlice() {
	// We must check if it is nil before checking length.
	// because the length of nil slice is 0.
	if d.value.IsNil() {
		d.printf("(%s)(nil)", d.value.Type().String())
		return
	}

	cleanup, ok := d.writeVisitedPointer()
	if ok {
		return
	}
	defer cleanup()

	d.writeArray()
}

func (d *dumper) writeArray() {
	if d.value.Len() == 0 {
		d.printf("%s{}", d.value.Type().String())
		return
	}
	d.writeRaw(d.value.Type().String())
	d.writeList()
}

func (d *dumper) writeList() {
	d.writeBlock(func() {
		typ := d.value.Type().Elem()
		size := 1
		if s, ok := d.listGroupingSize[typ]; ok && s > 1 {
			size = s
		}
		var breakLine bool
		for i := 0; i < d.value.Len() && d.err == nil; i++ {
			elem := d.value.Index(i)
			mod := (i + 1) % size
			breakLine = mod == 0
			if size == 1 || mod == 1 {
				d.writeIndent()
				d.writeValue(elem)
				d.writeRaw(",")
			} else {
				d.writeRaw(" ")
				d.writeValue(elem)
				d.writeRaw(",")
			}
			if breakLine {
				d.writeRaw("\n")
			}
		}
		if !breakLine {
			d.writeRaw("\n")
		}
	})
}

func (d *dumper) writeInterface() {
	elem := d.value.Elem()
	if elem.IsValid() {
		d.writeValue(elem)
		return
	}
	d.writeRaw("nil")
}

func (d *dumper) writeNumber() {
	switch d.value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		d.scratch = strconv.AppendInt(d.scratch[:0], d.value.Int(), 10)
		d.writeBytes(d.scratch)
		return
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		d.writeUnsignedInt()
		return
	case reflect.Float32, reflect.Float64:
		d.printf("%f", d.value.Float())
		return
	case reflect.Complex64:
		d.printf("%v", complex64(d.value.Complex()))
		return
	case reflect.Complex128:
		d.printf("%v", d.value.Complex())
		return
	}
	panic(fmt.Errorf("unreachable type: %s", d.value.Type()))
}

func (d *dumper) writeUnsignedInt() {
	switch d.value.Kind() {
	case reflect.Uint8:
		switch d.uintFormat {
		case BinaryUint:
			d.printf("0b%08b", d.value.Uint())
			return
		case HexUint:
			d.printf("0x%02x", d.value.Uint())
			return
		}
	case reflect.Uint16:
		switch d.uintFormat {
		case BinaryUint:
			d.printf("0b%016b", d.value.Uint())
			return
		case HexUint:
			d.printf("0x%04x", d.value.Uint())
			return
		}
	case reflect.Uint32:
		switch d.uintFormat {
		case BinaryUint:
			d.printf("0b%032b", d.value.Uint())
			return
		case HexUint:
			d.printf("0x%08x", d.value.Uint())
			return
		}
	case reflect.Uint64:
		switch d.uintFormat {
		case BinaryUint:
			d.printf("0b%064b", d.value.Uint())
			return
		case HexUint:
			d.printf("0x%016x", d.value.Uint())
			return
		}
	}
	d.scratch = strconv.AppendUint(d.scratch[:0], d.value.Uint(), 10)
	d.writeBytes(d.scratch)
}

func (d *dumper) writePointer() {
	d.printf(
		"(%s)(unsafe.Pointer(uintptr(0x%x)))",
		d.value.Type().String(),
		d.value.Pointer(),
	)
}

func (d *dumper) writeVisitedPointer() (func(), bool) {
	key := visit{typ: d.value.Type(), pointer: d.value.Pointer()}
	if d.value.Kind() == reflect.Slice {
		key.length = d.value.Len()
	}
	if d.visitPointers[key] {
		d.writePointer()
		return nil, true
	}
	d.visitPointers[key] = true
	return func() {
		delete(d.visitPointers, key)
	}, false
}

func (d *dumper) writeBlock(f func()) {
	d.writeRaw("{\n")
	if d.err != nil {
		return
	}
	d.depth++
	f()
	d.depth--
	d.writeIndentedRaw("}")
}

func (d *dumper) writeBool(b bool) {
	d.writeRaw(strconv.FormatBool(b))
}

func (d *dumper) writeString(s string) {
	d.writeRaw("\"")
	for len(s) > 0 && d.err == nil {
		n := len(s)
		if n > 4096 {
			n = 4096
			// Do not split a valid UTF-8 sequence across quoted chunks.
			for n > 0 && !utf8.RuneStart(s[n]) {
				n--
			}
			if n == 0 {
				// A run of invalid continuation bytes can be split anywhere.
				n = 4096
			}
		}
		d.scratch = strconv.AppendQuote(d.scratch[:0], s[:n])
		d.writeBytes(d.scratch[1 : len(d.scratch)-1])
		s = s[n:]
	}
	d.writeRaw("\"")
}

func (d *dumper) writeIndent() {
	if d.depth == 0 {
		return
	}
	// Write indentation in bounded chunks, even for very deep values.
	const spaces = "                                                                "
	for remaining := d.depth * d.indentSize; remaining > 0 && d.err == nil; {
		n := remaining
		if n > len(spaces) {
			n = len(spaces)
		}
		d.writeRaw(spaces[:n])
		remaining -= n
	}
}

func (d *dumper) writeIndentedRaw(s string) {
	d.writeIndent()
	d.writeRaw(s)
}

// writeRaw preserves the first output error.
func (d *dumper) writeRaw(s string) {
	if d.err != nil {
		return
	}
	n, err := io.WriteString(d.w, s)
	if err == nil && n != len(s) {
		err = io.ErrShortWrite
	}
	d.err = err
}

func (d *dumper) writeBytes(p []byte) {
	if d.err != nil {
		return
	}
	n, err := d.w.Write(p)
	if err == nil && n != len(p) {
		err = io.ErrShortWrite
	}
	d.err = err
}

func (d *dumper) printf(format string, a ...interface{}) {
	if d.err == nil {
		_, d.err = fmt.Fprintf(d.w, format, a...)
	}
}

type dumpWriter struct{ *dumper }

var _ Writer = (*dumpWriter)(nil)

func (d *dumpWriter) Write(s string) { d.dumper.writeRaw(s) }
func (d *dumpWriter) WriteBlock(s string) {
	d.dumper.writeRaw("{\n")
	d.dumper.depth++
	for s != "" && d.dumper.err == nil {
		line := s
		if i := strings.IndexByte(s, '\n'); i >= 0 {
			line, s = s[:i], s[i+1:]
		} else {
			s = ""
		}
		d.dumper.writeRaw(strings.Repeat("\t", d.dumper.depth))
		d.dumper.writeRaw(strings.TrimSuffix(line, "\r"))
		d.dumper.writeRaw("\n")
	}
	d.dumper.depth--
	d.dumper.writeRaw(strings.Repeat("\t", d.dumper.depth))
	d.dumper.writeRaw("}")
}
