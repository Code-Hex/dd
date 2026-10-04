//go:build go1.26
// +build go1.26

package dd_test

import (
	"strings"
	"testing"

	"github.com/Code-Hex/dd"
)

func TestGo126NewExpression(t *testing.T) {
	// new(expr) produces an ordinary pointer; retain the established pointer format.
	if got := dd.Dump(new(42)); !strings.HasPrefix(got, "(*int)(unsafe.Pointer(uintptr(") {
		t.Fatal(got)
	}
}
