package batch

import (
	"bufio"
	"io"
	"strings"
)

// ReadNumbers reads corporate numbers from a reader, one per line.
// Empty lines and lines starting with # are skipped.
func ReadNumbers(r io.Reader) ([]string, error) {
	var numbers []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		numbers = append(numbers, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return numbers, nil
}
