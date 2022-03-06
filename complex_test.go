package dd_test

import (
	"encoding/json"
	"go/parser"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	dd "github.com/Code-Hex/go-data-dumper"
	"github.com/google/go-cmp/cmp"
)

//go:generate go run cmd/wantdumper/main.go

var addressReplaceRegexp = regexp.MustCompile(`uintptr\((0x[\da-f]+)\)`)

func TestComplex(t *testing.T) {
	entries, err := os.ReadDir("testdata")
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		t.Run(entry.Name(), func(t *testing.T) {
			jsonFile := filepath.Join("testdata", entry.Name(), "data.json")
			content, err := ioutil.ReadFile(jsonFile)
			if err != nil {
				t.Fatal(err)
			}
			wantFile := filepath.Join("testdata", entry.Name(), "want.txt")
			want, err := ioutil.ReadFile(wantFile)
			if err != nil {
				t.Fatal(err)
			}
			var unmarshaled interface{}
			if err := json.Unmarshal(content, &unmarshaled); err != nil {
				t.Fatal(err)
			}

			got := dd.Dump(unmarshaled)
			// check syntax is valid
			if _, err := parser.ParseExpr(got); err != nil {
				t.Log(got)
				t.Fatal(err)
			}

			// replace addresses
			replacedWant := addressReplaceRegexp.ReplaceAll(want, []byte("0x0"))
			replacedGot := addressReplaceRegexp.ReplaceAllString(got, "0x0")

			if diff := cmp.Diff(string(replacedWant), replacedGot); diff != "" {
				t.Fatalf("(-want, +got)\n%s", diff)
			}
		})
	}
}
