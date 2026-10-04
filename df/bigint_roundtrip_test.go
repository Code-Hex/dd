package df_test

import (
	"bytes"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/Code-Hex/dd"
	"github.com/Code-Hex/dd/df"
)

func TestBigIntRoundTrip(t *testing.T) {
	huge, _ := new(big.Int).SetString("12345678901234567890123456789012345678901234567890", 10)
	values := []*big.Int{nil, big.NewInt(0), big.NewInt(-42), huge}
	var source strings.Builder
	source.WriteString("package main\nimport \"math/big\"\nfunc main() {\n")
	for i, value := range values {
		dump := dd.Dump(value, df.WithBigInt())
		var streamed bytes.Buffer
		if err := dd.DumpTo(&streamed, value, df.WithBigInt()); err != nil {
			t.Fatal(err)
		}
		if streamed.String() != dump {
			t.Fatal("DumpTo differs")
		}
		fmt.Fprintf(&source, "v%d := %s\n", i, dump)
		if value == nil {
			fmt.Fprintf(&source, "if v%d != nil { panic(\"nil lost\") }\n", i)
		} else {
			fmt.Fprintf(&source, "if v%d == nil || v%d.String() != %q { panic(\"value lost\") }\n", i, i, value.String())
		}
	}
	source.WriteString("}\n")
	path := filepath.Join(t.TempDir(), "main.go")
	if err := os.WriteFile(path, []byte(source.String()), 0600); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command(filepath.Join(runtime.GOROOT(), "bin", "go"), "run", path).CombinedOutput(); err != nil {
		t.Fatalf("generated code failed: %v\n%s\n%s", err, out, source.String())
	}
}
