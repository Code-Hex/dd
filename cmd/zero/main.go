package main

import (
	"bytes"
	"context"
	"fmt"
	"go/format"
	"os"
	"reflect"

	"github.com/Code-Hex/dd"
)

func main() {
	if err := run(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "err: %q", err.Error())
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	zeroPrimitives := []any{
		false,
		uint8(0),
		uint16(0),
		uint32(0),
		uint64(0),
		int8(0),
		int16(0),
		int32(0),
		int64(0),
		float32(0),
		float64(0),
		complex64(0),
		complex128(0),
		"",
		int(0),
		uint(0),
		uintptr(0),
	}
	var buf bytes.Buffer
	buf.WriteString("var zeroPrimitives = map[reflect.Type]string{\n")
	for _, zero := range zeroPrimitives {
		buf.WriteString("  ")
		rt := reflect.TypeOf(zero)
		fmt.Fprintf(&buf,
			"reflect.TypeOf(%s(%#v)): %q,\n",
			rt.String(), zero, dd.Dump(zero),
		)
	}
	buf.WriteString("}\n")

	src, err := format.Source(buf.Bytes())
	if err != nil {
		return err
	}

	f, err := os.Create("zero.go")
	if err != nil {
		return err
	}
	defer f.Close()
	f.WriteString("package dd\n\n")
	f.WriteString("import \"reflect\"\n\n")
	_, err = f.Write(src)
	return err
}
