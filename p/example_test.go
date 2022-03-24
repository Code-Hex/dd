package p_test

import (
	"bytes"
	"fmt"

	"github.com/Code-Hex/dd/p"
	"github.com/alecthomas/chroma/styles"
)

func ExampleP() {
	// prints simple string literal.
	p.P("Hello, World")
	// Output:
	// [38;5;186m"Hello, World"[0m
}

func ExamplePrinter_P() {
	// prints simple string literal.
	p.New(p.WithStyle(styles.Dracula)).P("Hello, World")
	// Output:
	// [38;5;228m"Hello, World"[0m
}

func ExampleFp() {
	// prints simple string literal.
	var buf bytes.Buffer
	p.Fp(&buf, "Hello, World")
	fmt.Print(buf.String())
	// Output:
	// [38;5;186m"Hello, World"[0m
}

func ExamplePrinter_Fp() {
	// prints simple string literal.
	var buf bytes.Buffer
	p.New(p.WithStyle(styles.Dracula)).Fp(&buf, "Hello, World")
	fmt.Print(buf.String())
	// Output:
	// [38;5;228m"Hello, World"[0m
}
