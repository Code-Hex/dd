package data

type OptionFunc func(*options)

func WithExportedOnly() OptionFunc {
	return func(o *options) {
		o.exportedOnly = true
	}
}

func WithIndent(indent int) OptionFunc {
	return func(o *options) {
		o.indentSize = indent
	}
}

func Dump(v interface{}, opts ...OptionFunc) string {
	return newDataDumper(v).build().String()
}
