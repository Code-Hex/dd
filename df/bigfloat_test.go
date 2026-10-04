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

func TestWithBigFloatRoundTrip(t *testing.T) {
	cases := []string{
		"(*big.Float)(nil)",
		"new(big.Float)",
		"new(big.Float).SetMode(big.ToNegativeInf)",
		"new(big.Float).SetFloat64(math.Copysign(0, -1)).SetPrec(0)",
		"new(big.Float).SetPrec(23).SetFloat64(math.Copysign(0, -1))",
		"new(big.Float).SetPrec(0).SetInf(false)",
		"new(big.Float).SetPrec(256).SetMode(big.AwayFromZero).SetInf(true)",
		"new(big.Float).SetFloat64(1.2345678901234567)",
	}
	for mode := big.ToNearestEven; mode <= big.ToPositiveInf; mode++ {
		cases = append(cases, fmt.Sprintf("func() *big.Float { v := new(big.Float).SetPrec(257).SetMode(big.%s); v.SetString(\"-1.234567890123456789012345678901234567890123456789\"); return v }()", mode))
	}
	for _, exponent := range []int{big.MinExp, big.MaxExp} {
		cases = append(cases, fmt.Sprintf("new(big.Float).SetPrec(128).SetMantExp(big.NewFloat(0.5), %d)", exponent))
	}
	// Generate values and check emitted expressions in separate programs so the
	// test verifies the public output as compilable, executable Go source.
	var source strings.Builder
	source.WriteString("package main\nimport (\"fmt\"; \"math\"; \"math/big\"; \"github.com/Code-Hex/dd\"; \"github.com/Code-Hex/dd/df\")\nfunc main() {\n")
	for _, value := range cases {
		fmt.Fprintf(&source, "fmt.Printf(\"%%s\\n---\\n\", dd.Dump(%s, df.WithBigFloat()))\n", value)
	}
	source.WriteString("}\n")
	dir := t.TempDir()
	input := filepath.Join(dir, "dump.go")
	if err := os.WriteFile(input, []byte(source.String()), 0600); err != nil {
		t.Fatal(err)
	}
	output, err := exec.Command(filepath.Join(runtime.GOROOT(), "bin", "go"), "run", input).CombinedOutput()
	if err != nil {
		t.Fatalf("generate expressions: %v\n%s", err, output)
	}
	dumps := strings.Split(strings.TrimSuffix(string(output), "\n---\n"), "\n---\n")
	if len(dumps) != len(cases) {
		t.Fatalf("got %d expressions, want %d", len(dumps), len(cases))
	}
	source.Reset()
	source.WriteString("package main\nimport (\"fmt\"; \"math\"; \"math/big\")\nfunc main() {\n")
	for i, value := range cases {
		fmt.Fprintf(&source, "{ want := %s; got := %s; if want == nil { if got != nil { panic(\"expected nil\") } } else if got == nil || got.Cmp(want) != 0 || got.Prec() != want.Prec() || got.Mode() != want.Mode() || got.Signbit() != want.Signbit() { panic(fmt.Sprintf(\"case %d: got %%v; want %%v\", got, want)) } }\n", value, dumps[i], i)
	}
	source.WriteString("}\n")
	check := filepath.Join(dir, "check.go")
	if err := os.WriteFile(check, []byte(source.String()), 0600); err != nil {
		t.Fatal(err)
	}
	if output, err := exec.Command(filepath.Join(runtime.GOROOT(), "bin", "go"), "run", check).CombinedOutput(); err != nil {
		t.Fatalf("round trip: %v\n%s", err, output)
	}
}

func TestWithBigFloatDumpTo(t *testing.T) {
	value, _ := new(big.Float).SetPrec(257).SetString("1.2345678901234567890123456789")
	var out bytes.Buffer
	if err := dd.DumpTo(&out, value, df.WithBigFloat()); err != nil {
		t.Fatal(err)
	}
	if want := dd.Dump(value, df.WithBigFloat()); out.String() != want {
		t.Fatalf("DumpTo = %q, want %q", out.String(), want)
	}
}
