package df_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Code-Hex/dd"
	"github.com/Code-Hex/dd/df"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestJSONRawMessageRoundTrip(t *testing.T) {
	cases := []json.RawMessage{nil, {}, []byte("{\"x\":\"a" + string(rune(96)) + "b\"}"), []byte("{\r\n\t\"x\":1\n}"), {0xff, 0, 0xc0}, []byte("{\"x\":\"日本語\"}")}
	var source strings.Builder
	source.WriteString("package main\nimport(\"bytes\";\"encoding/json\")\nfunc main(){\n")
	for i, value := range cases {
		dump := dd.Dump(value, df.WithJSONRawMessage())
		var out bytes.Buffer
		if err := dd.DumpTo(&out, value, df.WithJSONRawMessage()); err != nil {
			t.Fatal(err)
		}
		if dump != out.String() {
			t.Fatal("DumpTo differs")
		}
		fmt.Fprintf(&source, "v%d := %s\nif !bytes.Equal(v%d,[]byte(%q)) || (v%d==nil)!=%t {panic(\"bytes or nil changed\")}\n", i, dump, i, string(value), i, value == nil)
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
