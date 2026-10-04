//go:build go1.24
// +build go1.24

package dd_test

import (
	"testing"

	"github.com/Code-Hex/dd"
)

type boxAlias[T any] = genericBox[T]
type sliceAlias[T any] = []T

func TestGo124Aliases(t *testing.T) {
	checkDumpType(t, dd.Dump(boxAlias[int]{Value: 3}),
		"type genericBox[T any] struct { Value T }", "genericBox[int]")
	checkDumpType(t, dd.Dump(sliceAlias[int]{1, 2}), "", "[]int")
}
