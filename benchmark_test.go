package dd_test

import (
	"fmt"
	"go/parser"
	"go/token"
	"io"
	"strings"
	"testing"

	"github.com/Code-Hex/dd"
)

var benchmarkDump string

type benchmarkRecord struct {
	ID      int
	Name    string
	Enabled bool
	Values  []int
}
type benchmarkNode struct {
	Value int
	Next  *benchmarkNode
}

func BenchmarkDump(b *testing.B) {
	for _, size := range []int{100, 10000} {
		records := make([]benchmarkRecord, size)
		values := make(map[string]int, size)
		for i := range records {
			records[i] = benchmarkRecord{i, "example", true, []int{1, 2, 3}}
			values[fmt.Sprintf("key%06d", i)] = i
		}
		b.Run(fmt.Sprintf("Records/%d", size), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				benchmarkDump = dd.Dump(records)
			}
		})
		b.Run(fmt.Sprintf("Map/%d", size), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				benchmarkDump = dd.Dump(values)
			}
		})
	}
	for _, depth := range []int{64, 256} {
		var node *benchmarkNode
		for i := 0; i < depth; i++ {
			node = &benchmarkNode{Value: i, Next: node}
		}
		b.Run(fmt.Sprintf("Depth/%d", depth), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				benchmarkDump = dd.Dump(node)
			}
		})
	}
	var source strings.Builder
	source.WriteString("package example\n")
	for i := 0; i < 100; i++ {
		fmt.Fprintf(&source, "func F%d(x int) int { if x > 0 { return x + %d }; return 0 }\n", i, i)
	}
	file, err := parser.ParseFile(token.NewFileSet(), "example.go", source.String(), 0)
	if err != nil {
		b.Fatal(err)
	}
	b.Run("AST", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			benchmarkDump = dd.Dump(file, dd.WithOmitEmptyFields())
		}
	})
}

type plainDiscard struct{}

func (plainDiscard) Write(p []byte) (int, error) { return len(p), nil }

func BenchmarkWriterPaths(b *testing.B) {
	for _, writer := range []struct {
		name string
		w    io.Writer
	}{{"StringWriter", io.Discard}, {"Writer", plainDiscard{}}} {
		b.Run(writer.name, func(b *testing.B) {
			value := make([]int, 10000)
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				if err := dd.DumpTo(writer.w, value); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
	b.Run("LargeString", func(b *testing.B) {
		value := strings.Repeat("a", 1<<20)
		b.SetBytes(int64(len(value)))
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := dd.DumpTo(io.Discard, value); err != nil {
				b.Fatal(err)
			}
		}
	})
}
