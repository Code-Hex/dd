package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/Code-Hex/dd"
)

func main() {
	if err := run(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "err: %q", err.Error())
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	baseDir := "testdata"
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return fmt.Errorf("failed to read testdata: %w", err)
	}
	for _, entry := range entries {
		jsonFile := filepath.Join(baseDir, entry.Name(), "data.json")
		content, err := ioutil.ReadFile(jsonFile)
		if err != nil {
			return fmt.Errorf("failed to read %q: %w", jsonFile, err)
		}
		var unmarshaled interface{}
		if err := json.Unmarshal(content, &unmarshaled); err != nil {
			return fmt.Errorf("failed to unmarshal %q: %w", jsonFile, err)
		}
		dumped := dd.Dump(unmarshaled)
		wantFile := filepath.Join(baseDir, entry.Name(), "want.txt")
		if err := writeFile(wantFile, dumped); err != nil {
			return err
		}
	}
	return nil
}

func writeFile(filename, content string) error {
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create %q: %w", filename, err)
	}
	defer f.Close()

	if _, err := io.WriteString(f, content); err != nil {
		return fmt.Errorf("failed to write content %q: %w", filename, err)
	}
	return nil
}
