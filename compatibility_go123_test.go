//go:build go1.23
// +build go1.23

package dd_test

import (
	"iter"
	"strings"
	"testing"

	"github.com/Code-Hex/dd"
)

func TestGo123Iterators(t *testing.T) {
	seq := iter.Seq[int](func(func(int) bool) { panic("must not be called") })
	seq2 := iter.Seq2[string, int](func(func(string, int) bool) { panic("must not be called") })
	checkDumpType(t, strings.ReplaceAll(dd.Dump(seq), "iter.Seq[int]", "Seq[int]"),
		"type Seq[T any] func(func(T) bool)", "Seq[int]")
	checkDumpType(t, strings.ReplaceAll(dd.Dump(seq2), "iter.Seq2[string,int]", "Seq2[string,int]"),
		"type Seq2[K, V any] func(func(K, V) bool)", "Seq2[string,int]")
}
