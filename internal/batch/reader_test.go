package batch_test

import (
	"strings"
	"testing"

	"github.com/planitaicojp/houjin-cli/internal/batch"
)

func TestReadNumbers(t *testing.T) {
	input := "# comment\n1180301018771\n\n5180301018778\n"
	numbers, err := batch.ReadNumbers(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(numbers) != 2 {
		t.Fatalf("expected 2 numbers, got %d", len(numbers))
	}
	if numbers[0] != "1180301018771" {
		t.Errorf("expected 1180301018771, got %s", numbers[0])
	}
	if numbers[1] != "5180301018778" {
		t.Errorf("expected 5180301018778, got %s", numbers[1])
	}
}

func TestReadNumbers_empty(t *testing.T) {
	input := "# only comments\n\n"
	numbers, err := batch.ReadNumbers(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(numbers) != 0 {
		t.Errorf("expected 0 numbers, got %d", len(numbers))
	}
}

func TestReadNumbers_trimSpace(t *testing.T) {
	input := "  1180301018771  \n"
	numbers, err := batch.ReadNumbers(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(numbers) != 1 {
		t.Fatalf("expected 1 number, got %d", len(numbers))
	}
	if numbers[0] != "1180301018771" {
		t.Errorf("expected trimmed number, got %q", numbers[0])
	}
}
