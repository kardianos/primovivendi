package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
)

func main() {
	err := run()
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	const dir = "chapters"
	const outputFile = "book.md"

	// Read directory contents
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("Error reading directory: %v", err)
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		files = append(files, entry.Name())
	}
	sort.Strings(files)

	var out = &bytes.Buffer{}
	for i, file := range files {
		if i > 0 {
			_, err = out.WriteString("\n\n\\newpage\n\n")
			if err != nil {
				return fmt.Errorf("Error writing newline: %v", err)
			}
		}
		fn := filepath.Join(dir, file)
		in, err := os.Open(fn)
		if err != nil {
			return fmt.Errorf("Error opening file %s: %v", file, err)
		}

		_, err = io.Copy(out, in)
		in.Close()
		if err != nil {
			return fmt.Errorf("Error reading file %s: %v", file, err)
		}
	}

	return os.WriteFile(outputFile, out.Bytes(), 0600)
}
