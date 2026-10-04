//go:build go1.27
// +build go1.27

package dd_test

import (
	"testing"

	"github.com/Code-Hex/dd"
)

type methodHost struct{}

func (methodHost) Identity[T any](v T) T { return v }

type embeddedValue struct{ N int }
type outerValue struct{ embeddedValue }

func TestGo127Syntax(t *testing.T) {
	checkDumpType(t, dd.Dump(methodHost{}.Identity[int]), "", "func(int) int")
	value := outerValue{N: 42}
	checkDumpType(t, dd.Dump(value),
		"type embeddedValue struct { N int }; type outerValue struct { embeddedValue }", "outerValue")
	var f func(int) int
	f = identity
	checkDumpType(t, dd.Dump(f), "", "func(int) int")
}
