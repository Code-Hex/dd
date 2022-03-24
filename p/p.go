package p

import (
	"bytes"
	"io"

	"github.com/Code-Hex/dd"
	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters"
	"github.com/alecthomas/chroma/lexers/g"
	"github.com/alecthomas/chroma/styles"
	"github.com/mattn/go-colorable"
)

var (
	lexer          = g.Go // maybe returns error is nil if successed to compile regexp.
	defaultPrinter = New()
)

type options struct {
	ddOptions []dd.OptionFunc
	style     *chroma.Style
	formatter chroma.Formatter
}

func newOptions() *options {
	return &options{
		style:     styles.Monokai,
		formatter: formatters.TTY256, // Format method returns error is nil
	}
}

// Printer is a printer.
type Printer struct {
	options *options
}

// New creates a new Printer.
func New(opts ...OptionFunc) *Printer {
	o := newOptions()
	for _, optFunc := range opts {
		optFunc(o)
	}
	return &Printer{options: o}
}

// OptionFunc is type of an option for any printers.
type OptionFunc func(opts *options)

// WithDumpOptions is an option to append options of dd.Dump function.
func WithDumpOptions(ddOpts ...dd.OptionFunc) OptionFunc {
	return func(opts *options) {
		opts.ddOptions = append(opts.ddOptions, ddOpts...)
	}
}

// WithStyle is an option to set style of the syntax highlighting.
// Default will be set Monokai style.
//
// Available themes: https://pkg.go.dev/github.com/alecthomas/chroma/styles
func WithStyle(style *chroma.Style) OptionFunc {
	return func(opts *options) {
		opts.style = style
	}
}

// P prints dumped your specified data with colored.
// Spaces are always added between operands and a newline is appended.
// It returns the number of bytes written and any write error encountered.
func (p *Printer) P(args ...interface{}) (int, error) {
	return p.Fp(colorable.NewColorableStdout(), args...)
}

// Fp prints dumped your specified data with colored and writes to w.
// Spaces are always added between operands and a newline is appended.
// It returns the number of bytes written and any write error encountered.
func (p *Printer) Fp(w io.Writer, args ...interface{}) (int, error) {
	var buf bytes.Buffer
	for i, a := range args {
		if i > 0 {
			buf.WriteByte(' ')
		}
		dump := dd.Dump(a, p.options.ddOptions...)
		iterator, _ := lexer.Tokenise(nil, dump)
		p.options.formatter.Format(&buf, p.options.style, iterator)
	}
	buf.WriteByte('\n')
	cpn, cperr := io.Copy(w, &buf)
	return int(cpn), cperr
}

// P prints dumped your specified data with colored.
// Spaces are always added between operands and a newline is appended.
// It returns the number of bytes written and any write error encountered.
func P(args ...interface{}) (int, error) {
	return defaultPrinter.P(args...)
}

// Fp prints dumped your specified data with colored and writes to w.
// Spaces are always added between operands and a newline is appended.
// It returns the number of bytes written and any write error encountered.
func Fp(w io.Writer, args ...interface{}) (int, error) {
	return defaultPrinter.Fp(w, args...)
}
