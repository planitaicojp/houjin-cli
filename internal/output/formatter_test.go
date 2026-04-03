package output_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/planitaicojp/houjin-cli/internal/model"
	"github.com/planitaicojp/houjin-cli/internal/output"
)

func testResponse() *model.Response {
	return &model.Response{
		Count:        1,
		DivideNumber: 1,
		DivideSize:   1,
		Corporations: []model.Corporation{
			{
				CorporateNumber: "2180301018771",
				Name:            "トヨタ自動車株式会社",
				NameKana:        "トヨタジドウシャ",
				NameEnglish:     "TOYOTA MOTOR CORPORATION",
				Kind:            "301",
				Prefecture:      "愛知県",
				City:            "豊田市",
				Address:         "トヨタ町１番地",
				PostalCode:      "4718571",
			},
		},
	}
}

func TestJSONFormatter(t *testing.T) {
	var buf bytes.Buffer
	f := output.New("json")
	if err := f.Format(&buf, testResponse()); err != nil {
		t.Fatal(err)
	}

	var result model.Response
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if result.Count != 1 {
		t.Errorf("expected count 1, got %d", result.Count)
	}
	if result.Corporations[0].Name != "トヨタ自動車株式会社" {
		t.Errorf("unexpected name: %s", result.Corporations[0].Name)
	}
}

func TestTableFormatter(t *testing.T) {
	var buf bytes.Buffer
	f := output.New("table")
	if err := f.Format(&buf, testResponse()); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "2180301018771") {
		t.Error("table output should contain corporate number")
	}
	if !strings.Contains(out, "トヨタ自動車株式会社") {
		t.Error("table output should contain name")
	}
}

func TestCSVFormatter(t *testing.T) {
	var buf bytes.Buffer
	f := output.New("csv")
	if err := f.Format(&buf, testResponse()); err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines (header + 1 row), got %d", len(lines))
	}
	if !strings.HasPrefix(lines[0], "corporate_number,") {
		t.Errorf("expected CSV header, got: %s", lines[0])
	}
	if !strings.HasPrefix(lines[1], "2180301018771,") {
		t.Errorf("expected data row, got: %s", lines[1])
	}
}

func TestNewFormatter_default(t *testing.T) {
	f := output.New("unknown")
	var buf bytes.Buffer
	if err := f.Format(&buf, testResponse()); err != nil {
		t.Fatal(err)
	}
	var result model.Response
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("default formatter should produce valid JSON: %v", err)
	}
}
