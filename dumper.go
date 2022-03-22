package dd

import (
	"bufio"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"text/tabwriter"

	"github.com/Code-Hex/dd/internal/sort"
)

type dumpFunc func(reflect.Value, Writer)

type options struct {
	exportedOnly     bool
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

type dumper struct {
	buf              *strings.Builder
	tw               *tabwriter.Writer
	value            reflect.Value
	depth            int
	visitPointers    map[uintptr]bool
	cachedZeroValues map[reflect.Type]string
	clonePool        sync.Pool
	// options
	exportedOnly     bool
	uintFormat       UintFormat
	convertibleTypes map[reflect.Type]dumpFunc
	listGroupingSize map[reflect.Type]int
}

var _ interface {
	fmt.Stringer
} = (*dumper)(nil)

func newDataDumper(obj interface{}, optFuncs ...OptionFunc) *dumper {
	opts := newDefaultOptions()
	// apply options
	for _, apply := range optFuncs {
		apply(opts)
	}
	clonePool := sync.Pool{
		New: func() interface{} {
			buf := new(strings.Builder)
			return &dumper{
				buf: buf,
				tw:  tabwriter.NewWriter(buf, opts.indentSize, 0, 1, ' ', 0),
			}
		},
	}
	ret := clonePool.Get().(*dumper)
	ret.value = valueOf(obj, true)
	ret.clonePool = clonePool
	ret.visitPointers = make(map[uintptr]bool)
	ret.cachedZeroValues = zeroPrimitives
	ret.clonePool = clonePool
	ret.exportedOnly = opts.exportedOnly
	ret.uintFormat = opts.uintFormat
	ret.convertibleTypes = opts.convertibleTypes
	ret.listGroupingSize = opts.listGroupingSize
	return ret
}

func (d *dumper) clone(obj interface{}) *dumper {
	child := d.clonePool.Get().(*dumper)
	child.value = valueOf(obj, false)
	child.depth = d.depth
	child.visitPointers = d.visitPointers
	child.cachedZeroValues = d.cachedZeroValues
	child.clonePool = d.clonePool
	child.exportedOnly = d.exportedOnly
	child.uintFormat = d.uintFormat
	child.convertibleTypes = d.convertibleTypes
	child.listGroupingSize = d.listGroupingSize
	return child
}

func dumpclone(d *dumper, obj interface{}) string {
	cloned := d.clone(obj).build()
	ret := cloned.String()
	cloned.release()
	return ret
}

func (d *dumper) release() {
	d.buf.Reset()
	d.clonePool.Put(d)
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
	if kind == reflect.Invalid {
		d.writeRaw("nil")
		return d
	}

	convertFunc, ok := d.convertibleTypes[d.value.Type()]
	if ok {
		convertFunc(d.value, &dumpWriter{d})
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

	if ok := d.writeVisitedPointer(); ok {
		return
	}

	d.writeRaw(d.value.Type().String())
	d.writeRaw(" ")

	d.writeBlock(func() {
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
			zeroValues = append(zeroValues, d.zeroValue(outTyp))
		}
		d.indentedPrintf("return %s\n", strings.Join(zeroValues, ", "))
	})
}

//go:generate go run cmd/zero/main.go

func (d *dumper) zeroValue(rt reflect.Type) string {
	if cached, ok := d.cachedZeroValues[rt]; ok {
		return cached
	}
	zero := dumpclone(d, reflect.Zero(rt))
	d.cachedZeroValues[rt] = zero
	return zero
}

func (d *dumper) writePtr() {
	if d.value.IsNil() {
		d.printf("(%s)(nil)", d.value.Type())
		return
	}
	if ok := d.writeVisitedPointer(); ok {
		return
	}

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
	convertFunc, ok := d.convertibleTypes[deref.Type()]
	if ok {
		convertFunc(d.value, &dumpWriter{d})
		return
	}
	d.printf("&%s", dumpclone(d, deref))
	return
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
			d.indentedPrintf("%s: %s,\n", field.Name, dumpclone(d, fieldVal))
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

	if ok := d.writeVisitedPointer(); ok {
		return
	}

	d.writeRaw(d.value.Type().String())

	d.writeBlock(func() {
		keys := sort.Keys(d.value.MapKeys())
		for _, key := range keys {
			val := d.value.MapIndex(key)
			d.indentedPrintf("%s:\t%s,\n",
				dumpclone(d, key),
				dumpclone(d, val),
			)
		}
	})
}

func (d *dumper) writeSlice() {
	// We must check if it is nil before checking length.
	// because the length of nil slice is 0.
	if d.value.IsNil() {
		d.printf("(%s)(nil)", d.value.Type().String())
		return
	}

	if ok := d.writeVisitedPointer(); ok {
		return
	}

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
		for i := 0; i < d.value.Len(); i++ {
			elem := d.value.Index(i)
			mod := (i + 1) % size
			breakLine = mod == 0
			if size == 1 || mod == 1 {
				d.indentedPrintf("%s,", dumpclone(d, elem))
			} else {
				d.printf(" %s,", dumpclone(d, elem))
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
		d.writeRaw(dumpclone(d, elem))
		return
	}
	d.writeRaw("nil")
}

func (d *dumper) writeNumber() {
	switch d.value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		d.printf("%d", d.value.Int())
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
	d.writeRaw(strconv.FormatUint(d.value.Uint(), 10))
	return
}

func (d *dumper) writePointer() {
	d.printf(
		"(%s)(unsafe.Pointer(uintptr(0x%x)))",
		d.value.Type().String(),
		d.value.Pointer(),
	)
	return
}

func (d *dumper) writeVisitedPointer() bool {
	pointer := d.value.Pointer()
	if d.visitPointers[pointer] {
		d.writePointer()
		return true
	}
	d.visitPointers[pointer] = true
	return false
}

func (d *dumper) writeBlock(f func()) {
	d.writeRaw("{\n")
	d.depth++
	f()
	d.depth--
	d.writeIndentedRaw("}")
}

func (d *dumper) writeBool(b bool) {
	d.writeRaw(strconv.FormatBool(b))
}

func (d *dumper) writeString(s string) {
	d.writeRaw(strconv.Quote(s))
}

func (d *dumper) writeIndent() {
	if d.depth == 0 {
		return
	}
	d.writeRaw(d.indent())
}

func (d *dumper) writeIndentedRaw(s string) {
	d.writeIndent()
	d.writeRaw(s)
}

func (d *dumper) indentedPrintf(format string, a ...interface{}) {
	d.writeIndent()
	d.printf(format, a...)
}

// writeRaw appends the contents of s to p's buffer.
func (d *dumper) writeRaw(s string) {
	io.WriteString(d.tw, s)
}

func (d *dumper) printf(format string, a ...interface{}) {
	fmt.Fprintf(d.tw, format, a...)
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
