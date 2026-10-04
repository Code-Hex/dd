package p

import (
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

// WithFormatter is an option to set formatter of the terminal output.
// Default will be set TTY256(256-colour) formatter.
//
// Available chroma built-in formatters: https://pkg.go.dev/github.com/alecthomas/chroma/formatters
func WithFormatter(formatter chroma.Formatter) OptionFunc {
	return func(opts *options) {
		opts.formatter = formatter
	}
}

// P prints dumped your specified data with colored.
// Spaces are always added between operands and a newline is appended.
// It returns the number of bytes written and any formatting or write error encountered.
func (p *Printer) P(args ...interface{}) (int, error) {
	return p.Fp(colorable.NewColorableStdout(), args...)
}

// Fp prints dumped your specified data with colored and writes to w.
// Spaces are added between operands and a newline is appended on success.
// It returns the number of bytes written and the first formatting or write error.
// Output may be partial on error. Each operand is buffered for tokenization,
// but formatted output is written directly to w.
func (p *Printer) Fp(w io.Writer, args ...interface{}) (int, error) {
	out := &countingWriter{writer: w}
	for i, a := range args {
		if i > 0 {
			if _, err := io.WriteString(out, " "); err != nil {
				return out.n, err
			}
		}
		dump := dd.Dump(a, p.options.ddOptions...)
		iterator, err := lexer.Tokenise(nil, dump)
		if err != nil {
			return out.n, err
		}
		err = p.options.formatter.Format(out, p.options.style, iterator)
		// Some Chroma formatters ignore write errors.
		if out.err != nil {
			return out.n, out.err
		}
		if err != nil {
			return out.n, err
		}
	}
	_, err := io.WriteString(out, "\n")
	return out.n, err
}

type countingWriter struct {
	writer io.Writer
	n      int
	err    error
}

func (w *countingWriter) Write(b []byte) (int, error) {
	if w.err != nil {
		return 0, w.err
	}
	n, err := w.writer.Write(b)
	w.n += n
	if err == nil && n < len(b) {
		err = io.ErrShortWrite
	}
	w.err = err
	return n, err
}

// P prints dumped your specified data with colored.
// Spaces are always added between operands and a newline is appended.
// It returns the number of bytes written and any formatting or write error encountered.
func P(args ...interface{}) (int, error) {
	return defaultPrinter.P(args...)
}

// Fp prints dumped your specified data with colored and writes to w.
// Spaces are added between operands and a newline is appended on success.
// It returns the number of bytes written and the first formatting or write error.
// Output may be partial on error. Each operand is buffered for tokenization,
// but formatted output is written directly to w.
func Fp(w io.Writer, args ...interface{}) (int, error) {
	return defaultPrinter.Fp(w, args...)
}
