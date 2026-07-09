package md

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// LoadVars reads the YAML front matter from the first sorted chapter file
// (typically 0000_title_block.md) and returns string variables for {{.key}} expansion.
func LoadVars(chapterDir string) (map[string]string, error) {
	files, err := ListMarkdown(chapterDir)
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return map[string]string{}, nil
	}
	data, err := os.ReadFile(filepath.Join(chapterDir, files[0]))
	if err != nil {
		return nil, err
	}
	return ParseFrontMatterVars(data)
}

// ParseFrontMatterVars extracts simple key: value pairs from a YAML front matter block.
func ParseFrontMatterVars(data []byte) (map[string]string, error) {
	vars := make(map[string]string)
	s := string(data)
	if !strings.HasPrefix(strings.TrimSpace(s), "---") {
		return vars, nil
	}
	// Find closing ---
	rest := s
	if i := strings.Index(s, "---"); i >= 0 {
		rest = s[i+3:]
	}
	end := strings.Index(rest, "\n---")
	if end < 0 {
		return vars, nil
	}
	fm := rest[:end]

	re := regexp.MustCompile(`^([A-Za-z_][A-Za-z0-9_]*)\s*:\s*(.*)$`)
	sc := bufio.NewScanner(strings.NewReader(fm))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// skip YAML list / nested geometry etc.
		if strings.HasPrefix(line, "-") {
			continue
		}
		m := re.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		key := m[1]
		val := strings.TrimSpace(m[2])
		val = strings.Trim(val, `"'`)
		// only keep simple scalar values (not nested maps)
		if strings.HasPrefix(val, "[") || strings.HasPrefix(val, "{") {
			continue
		}
		vars[key] = val
	}
	return vars, sc.Err()
}

var tmplVar = regexp.MustCompile(`\{\{\s*\.([A-Za-z_][A-Za-z0-9_]*)\s*\}\}`)

// ExpandVars replaces {{.key}} with values from vars. Unknown keys are left empty
// (matching Go template zero-value behavior for missing map keys when using a map).
func ExpandVars(text string, vars map[string]string) string {
	return tmplVar.ReplaceAllStringFunc(text, func(m string) string {
		sub := tmplVar.FindStringSubmatch(m)
		if sub == nil {
			return m
		}
		if v, ok := vars[sub[1]]; ok {
			return v
		}
		// also try Title-case key used in some templates
		return ""
	})
}

// StripFrontMatter removes a leading YAML front matter block.
func StripFrontMatter(data []byte) []byte {
	b := bytes.TrimSpace(data)
	if !bytes.HasPrefix(b, []byte("---")) {
		return data
	}
	// after opening ---
	rest := b[3:]
	if i := bytes.Index(rest, []byte("\n---")); i >= 0 {
		after := rest[i+4:] // skip \n---
		if len(after) > 0 && after[0] == '\n' {
			after = after[1:]
		}
		return after
	}
	return data
}

// ListMarkdown returns sorted .md filenames in dir, excluding *_exp.md.
func ListMarkdown(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", dir, err)
	}
	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".md") {
			continue
		}
		if strings.HasSuffix(name, "_exp.md") {
			continue
		}
		files = append(files, name)
	}
	// ReadDir is sorted; keep stable order
	return files, nil
}
