//go:build go1.18
// +build go1.18

package dd_test

import (
	"testing"

	"github.com/Code-Hex/dd"
)

type genericBox[T any] struct{ Value T }
type genericFunc[T any] func(...T) T

func identity[T any](v T) T { return v }

func TestGo118Syntax(t *testing.T) {
	checkDumpType(t, dd.Dump(genericBox[int]{Value: 42}),
		"type genericBox[T any] struct { Value T }", "genericBox[int]")
	checkDumpType(t, dd.Dump(genericFunc[int](func(...int) int { return 1 })),
		"type genericFunc[T any] func(...T) T", "genericFunc[int]")
	checkDumpType(t, dd.Dump(identity[int]), "", "func(int) int")
}
