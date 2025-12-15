package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"unicode"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	const dir = "chapters"
	const outputFile = "book.md"

	// Read and sort chapter files
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("reading directory: %w", err)
	}

	var files []string
	for _, e := range entries {
		if e.IsDir() || strings.HasSuffix(e.Name(), "_exp.md") {
			continue
		}
		files = append(files, e.Name())
	}
	sort.Strings(files)

	if len(files) == 0 {
		return fmt.Errorf("no markdown files found in %s", dir)
	}

	// Create a template with the header data available as {{.Field}}
	tmpl := template.New("book").Delims("{{", "}}")

	// We'll parse each chapter with this template so all have access to the vars
	// We use a FuncMap if you want extra functions later
	tmpl, err = tmpl.Funcs(template.FuncMap{
			// you can add helpers here, e.g. "upper", "date", etc.
	}).Parse("") // start empty, we'll parse files one by one
	if err != nil {
		return err
	}

	var out bytes.Buffer
	var m map[string]any

	for i, fname := range files {
		if i > 0 {
			out.WriteString("\n\n\\newpage\n\n")
		}

		fullPath := filepath.Join(dir, fname)

		// Read the entire file content
		content, err := os.ReadFile(fullPath)
		if err != nil {
			return fmt.Errorf("reading %s: %w", fname, err)
		}
		if i == 0 {
			m, err = ParseYAML(content)
			if err != nil {
				return fmt.Errorf("parse yaml yeader")
			}
		}

		// Parse the chapter content as a template (so {{.Title}} etc. work)
		chapterTmpl, err := tmpl.Clone()
		if err != nil {
			return err
		}
		chapterTmpl, err = chapterTmpl.Parse(string(content))
		if err != nil {
			return fmt.Errorf("parsing template in %s: %w", fname, err)
		}

		// Execute the template with the header data
		if err := chapterTmpl.Execute(&out, m); err != nil {
			return fmt.Errorf("executing template %s: %w", fname, err)
		}
	}

	return os.WriteFile(outputFile, out.Bytes(), 0644)
}

// ParseYAML parses a simple subset of YAML into map[string]any
// Supports:
//
//	key: value
//	key: "string with spaces"
//	key: 123
//	key: true/false
//	nested: {a: 1, b: 2}
//	lists: [a, b, c]
//	multi-line strings with |
//	basic indentation (2 or 4 spaces)
func ParseYAML(data []byte) (map[string]any, error) {
	result := make(map[string]any)
	scanner := bufio.NewScanner(bytes.NewReader(data))

	var indentStack []int
	currentIndent := 0
	currentMap := result
	var mapStack []map[string]any
	var keyStack []string

	lineNum := 0
	reKeyValue := regexp.MustCompile(`^(\s*)([^:#]+?)\s*:\s*(.*)$`)
	reListItem := regexp.MustCompile(`^(\s*)-\s+(.*)$`)
	reFlowMap := regexp.MustCompile(`^\s*\{(.*)\}\s*$`)
	reFlowList := regexp.MustCompile(`^\s*\[(.*)\]\s*$`)

	for scanner.Scan() {
		lineNum++
		rawLine := scanner.Text()
		line := strings.TrimRightFunc(rawLine, unicode.IsSpace)

		if line == "" || strings.HasPrefix(line, "#") {
			continue // empty or comment
		}

		// Detect document separator
		if strings.TrimSpace(line) == "..." || strings.TrimSpace(line) == "---" {
			continue
		}

		// Calculate indent
		indent := len(rawLine) - len(strings.TrimLeft(rawLine, " "))
		//spaces := strings.Repeat(" ", indent)

		// Handle block scalars (e.g. abstract: |)
		if strings.HasSuffix(line, "|") || strings.HasSuffix(line, ">") {
			key := strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(line), "|"))
			key = strings.TrimSpace(strings.TrimSuffix(key, ">"))
			value, err := readBlockScalar(scanner, indent)
			if err != nil {
				return nil, fmt.Errorf("line %d: %w", lineNum, err)
			}
			currentMap[key] = value
			continue
		}

		// Flow style map: key: {a: 1, b: 2}
		if m := reFlowMap.FindStringSubmatch(line); m != nil {
			inner, _ := parseFlowMap(m[1])
			kv := reKeyValue.FindStringSubmatch(rawLine)
			if len(kv) > 2 {
				currentMap[strings.TrimSpace(kv[2])] = inner
			}
			continue
		}

		// Flow style list: key: [a, b, c]
		if m := reFlowList.FindStringSubmatch(line); m != nil {
			items := parseFlowList(m[1])
			kv := reKeyValue.FindStringSubmatch(rawLine)
			if len(kv) > 2 {
				currentMap[strings.TrimSpace(kv[2])] = items
			}
			continue
		}

		// Regular key: value
		if matches := reKeyValue.FindStringSubmatch(rawLine); len(matches) > 3 {
			key := strings.TrimSpace(matches[2])
			valueStr := strings.TrimSpace(matches[3])

			// Remove trailing comment
			if idx := strings.Index(valueStr, " #"); idx != -1 {
				valueStr = strings.TrimSpace(valueStr[:idx])
			}

			value := parseValue(valueStr)

			// Handle indentation changes
			for len(indentStack) > 0 && indent <= indentStack[len(indentStack)-1] {
				// Pop stack
				indentStack = indentStack[:len(indentStack)-1]
				mapStack = mapStack[:len(mapStack)-1]
				keyStack = keyStack[:len(keyStack)-1]
				currentMap = result
				if len(mapStack) > 0 {
					currentMap = mapStack[len(mapStack)-1]
				}
			}

			if indent > currentIndent && len(keyStack) > 0 {
				// Enter nested map
				parentKey := keyStack[len(keyStack)-1]
				nested := make(map[string]any)
				currentMap[parentKey] = nested
				currentMap = nested
				mapStack = append(mapStack, currentMap)
				keyStack = append(keyStack, key)
				indentStack = append(indentStack, currentIndent)
			}

			currentMap[key] = value
			currentIndent = indent
			if valueStr == "" {
				// This key will get a nested object
				keyStack = append(keyStack, key)
				nested := make(map[string]any)
				currentMap[key] = nested
				currentMap = nested
				mapStack = append(mapStack, currentMap)
				indentStack = append(indentStack, indent)
			}
		} else if matches := reListItem.FindStringSubmatch(rawLine); len(matches) > 2 {
			// Very basic list support under current map
			key := "_list_" + strconv.Itoa(lineNum)
			currentMap[key] = parseValue(matches[2])
		}
	}

	return result, scanner.Err()
}

func parseValue(s string) any {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	if s == "true" {
		return true
	}
	if s == "false" {
		return false
	}
	if s == "null" || s == "~" {
		return nil
	}
	if strings.HasPrefix(s, `"`) && strings.HasSuffix(s, `"`) {
		return strings.Trim(s, `"`)
	}
	if strings.HasPrefix(s, `'`) && strings.HasSuffix(s, `'`) {
		return strings.Trim(s, `'`)
	}
	if i, err := strconv.Atoi(s); err == nil {
		return i
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}
	return s
}

func readBlockScalar(scanner *bufio.Scanner, baseIndent int) (string, error) {
	var lines []string
	for scanner.Scan() {
		line := scanner.Text()
		indent := len(line) - len(strings.TrimLeft(line, " "))
		if indent <= baseIndent && strings.TrimSpace(line) != "" {
			// End of block scalar
			break
		}
		if indent > baseIndent {
			lines = append(lines, line[baseIndent+1:])
		} else if strings.TrimSpace(line) != "" {
			lines = append(lines, line)
		}
	}
	return strings.Join(lines, "\n"), nil
}

func parseFlowMap(s string) (map[string]any, error) {
	m := make(map[string]any)
	parts := splitFlow(s, ',')
	for _, p := range parts {
		if kv := strings.SplitN(p, ":", 2); len(kv) == 2 {
			k := strings.TrimSpace(kv[0])
			k = strings.Trim(k, `"'`)
			v := parseValue(kv[1])
			m[k] = v
		}
	}
	return m, nil
}

func parseFlowList(s string) []any {
	var list []any
	parts := splitFlow(s, ',')
	for _, p := range parts {
		list = append(list, parseValue(p))
	}
	return list
}

func splitFlow(s string, sep rune) []string {
	var parts []string
	var current strings.Builder
	bracket := 0
	for _, ch := range s {
		if ch == '{' || ch == '[' {
			bracket++
		} else if ch == '}' || ch == ']' {
			bracket--
		}
		if ch == sep && bracket == 0 {
			parts = append(parts, strings.TrimSpace(current.String()))
			current.Reset()
		} else {
			current.WriteRune(ch)
		}
	}
	if current.Len() > 0 {
		parts = append(parts, strings.TrimSpace(current.String()))
	}
	return parts
}
