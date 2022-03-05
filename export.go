package data

type OptionFunc func(*dumper)

func WithExportedOnly() OptionFunc {
	return func(d *dumper) {
		d.exportedOnly = true
	}
}

func Dump(v interface{}, opts ...OptionFunc) string {
	return newDataDumper(v).build().String()
}
