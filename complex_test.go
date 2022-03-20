package dd_test

import (
	"encoding/json"
	"fmt"
	"go/parser"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/Code-Hex/dd"
	"github.com/google/go-cmp/cmp"
)

//go:generate go run cmd/wantdumper/main.go

var addressReplaceRegexp = regexp.MustCompile(`uintptr\((0x[\da-f]+)\)`)

type Entry struct {
	name  string
	value any
	want  []byte
}

func makeEntries() ([]Entry, error) {
	entries, err := os.ReadDir("testdata")
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}
	ret := make([]Entry, len(entries))
	for i, entry := range entries {
		jsonFile := filepath.Join("testdata", entry.Name(), "data.json")
		content, err := ioutil.ReadFile(jsonFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %q: %w", jsonFile, err)
		}
		wantFile := filepath.Join("testdata", entry.Name(), "want.txt")
		want, err := ioutil.ReadFile(wantFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %q: %w", wantFile, err)
		}
		var unmarshaled any
		if err := json.Unmarshal(content, &unmarshaled); err != nil {
			return nil, fmt.Errorf("failed to unmarshal: %w", err)
		}
		ret[i] = Entry{
			name:  entry.Name(),
			value: unmarshaled,
			want:  want,
		}
	}
	return ret, nil
}

func TestComplex(t *testing.T) {
	entries, err := makeEntries()
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		entry := entry
		t.Run(entry.name, func(t *testing.T) {
			got := dd.Dump(entry.value)
			// check syntax is valid
			if _, err := parser.ParseExpr(got); err != nil {
				t.Log(got)
				t.Fatal(err)
			}

			// replace addresses
			replacedWant := addressReplaceRegexp.ReplaceAll(entry.want, []byte("0x0"))
			replacedGot := addressReplaceRegexp.ReplaceAllString(got, "0x0")

			if diff := cmp.Diff(string(replacedWant), replacedGot); diff != "" {
				t.Fatalf("(-want, +got)\n%s", diff)
			}
		})
	}
}

// 2022-03-20
// goos: darwin
// goarch: arm64
// pkg: github.com/Code-Hex/dd
// BenchmarkComplex/simple-8         	   34816	     32958 ns/op	   31119 B/op	     664 allocs/op
// BenchmarkComplex/twitter-search-adaptive-8         	      19	  59854191 ns/op	68459998 B/op	  599905 allocs/op
// PASS
// ok  	github.com/Code-Hex/dd	3.076s

func BenchmarkComplex(b *testing.B) {
	entries, err := makeEntries()
	if err != nil {
		b.Fatal(err)
	}
	for _, entry := range entries {
		entry := entry
		b.Run(entry.name, func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				dd.Dump(entry.value)
			}
		})
	}
}
