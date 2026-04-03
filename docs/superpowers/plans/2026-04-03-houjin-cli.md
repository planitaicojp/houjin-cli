# houjin-cli Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a Go CLI tool that wraps the National Tax Agency's Corporate Number Web-API (法人番号システム Web-API) with JSON output, config management, and check-digit validation.

**Architecture:** Cobra-based CLI with three subcommands (`get`, `search`, `diff`) mapping to API endpoints `/4/num`, `/4/name`, `/4/diff`. Internal packages handle XML→struct parsing, config (YAML + env override), multi-format output, and typed errors with exit codes. Follows conoha-cli patterns.

**Tech Stack:** Go 1.26+, spf13/cobra, gopkg.in/yaml.v3, Go stdlib (`encoding/xml`, `encoding/json`, `encoding/csv`, `net/http`, `text/tabwriter`)

---

## File Map

| File | Responsibility |
|------|---------------|
| `main.go` | Entry point — calls `cmd.Execute()` |
| `go.mod` | Module `github.com/planitaicojp/houjin-cli` |
| `cmd/root.go` | Root command, global flags (`--format`, `--verbose`, `--config`), `Execute()` |
| `cmd/get.go` | `houjin get <number...>` subcommand |
| `cmd/search.go` | `houjin search <name>` subcommand |
| `cmd/diff.go` | `houjin diff --from --to` subcommand |
| `cmd/version.go` | `houjin version` subcommand |
| `cmd/completion.go` | `houjin completion` subcommand |
| `internal/errors/errors.go` | Error types (APIError, ConfigError, ValidationError) + exit codes |
| `internal/model/corporation.go` | Corporation, Response structs + XML tags + check-digit validation |
| `internal/config/config.go` | YAML config load from `~/.config/houjin/config.yaml` |
| `internal/config/env.go` | Environment variable constants + `EnvOr` helper |
| `internal/output/formatter.go` | Formatter interface + `New()` factory |
| `internal/output/json.go` | JSON formatter |
| `internal/output/table.go` | Table formatter |
| `internal/output/csv.go` | CSV formatter |
| `internal/api/client.go` | HTTP client — base URL, app ID injection, XML fetch + parse |
| `internal/api/houjin.go` | `GetByNumber`, `SearchByName`, `GetDiff` methods |
| `testdata/num_response.xml` | Sample XML for `/4/num` endpoint |
| `testdata/name_response.xml` | Sample XML for `/4/name` endpoint |
| `testdata/diff_response.xml` | Sample XML for `/4/diff` endpoint |

---

## Task 1: Project Scaffold + Errors Package

**Files:**
- Create: `go.mod`, `main.go`, `internal/errors/errors.go`, `internal/errors/errors_test.go`

- [ ] **Step 1: Initialize Go module**

```bash
cd /root/dev/planitai/houjin-cli
go mod init github.com/planitaicojp/houjin-cli
```

- [ ] **Step 2: Write errors test**

Create `internal/errors/errors_test.go`:

```go
package errors_test

import (
	"testing"

	cerrors "github.com/planitaicojp/houjin-cli/internal/errors"
)

func TestAPIError(t *testing.T) {
	err := &cerrors.APIError{StatusCode: 403, Message: "rate limited"}
	if err.Error() != "API error (HTTP 403): rate limited" {
		t.Errorf("unexpected error: %s", err.Error())
	}
	if err.ExitCode() != cerrors.ExitAPI {
		t.Errorf("expected exit code %d, got %d", cerrors.ExitAPI, err.ExitCode())
	}
}

func TestConfigError(t *testing.T) {
	err := &cerrors.ConfigError{Message: "app_id not set"}
	if err.Error() != "config error: app_id not set" {
		t.Errorf("unexpected error: %s", err.Error())
	}
	if err.ExitCode() != cerrors.ExitConfig {
		t.Errorf("expected exit code %d, got %d", cerrors.ExitConfig, err.ExitCode())
	}
}

func TestValidationError(t *testing.T) {
	err := &cerrors.ValidationError{Field: "corporate_number", Message: "invalid check digit"}
	if err.Error() != "validation error on corporate_number: invalid check digit" {
		t.Errorf("unexpected error: %s", err.Error())
	}
	if err.ExitCode() != cerrors.ExitValidation {
		t.Errorf("expected exit code %d, got %d", cerrors.ExitValidation, err.ExitCode())
	}
}

func TestGetExitCode_nil(t *testing.T) {
	if cerrors.GetExitCode(nil) != cerrors.ExitOK {
		t.Error("expected ExitOK for nil error")
	}
}

func TestGetExitCode_generic(t *testing.T) {
	err := fmt.Errorf("generic")
	if cerrors.GetExitCode(err) != cerrors.ExitGeneral {
		t.Error("expected ExitGeneral for generic error")
	}
}
```

- [ ] **Step 3: Run test to verify it fails**

```bash
cd /root/dev/planitai/houjin-cli && go test ./internal/errors/...
```

Expected: compilation error — package doesn't exist yet.

- [ ] **Step 4: Implement errors package**

Create `internal/errors/errors.go`:

```go
package errors

import "fmt"

const (
	ExitOK         = 0
	ExitGeneral    = 1
	ExitConfig     = 2
	ExitAPI        = 3
	ExitValidation = 4
)

// ExitCoder is implemented by errors that carry a process exit code.
type ExitCoder interface {
	ExitCode() int
}

// APIError represents an error returned by the 法人番号 API.
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error (HTTP %d): %s", e.StatusCode, e.Message)
}

func (e *APIError) ExitCode() int { return ExitAPI }

// ConfigError represents a configuration problem.
type ConfigError struct {
	Message string
}

func (e *ConfigError) Error() string {
	return fmt.Sprintf("config error: %s", e.Message)
}

func (e *ConfigError) ExitCode() int { return ExitConfig }

// ValidationError represents invalid user input.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("validation error on %s: %s", e.Field, e.Message)
	}
	return fmt.Sprintf("validation error: %s", e.Message)
}

func (e *ValidationError) ExitCode() int { return ExitValidation }

// GetExitCode returns the exit code for the given error.
func GetExitCode(err error) int {
	if err == nil {
		return ExitOK
	}
	if ec, ok := err.(ExitCoder); ok {
		return ec.ExitCode()
	}
	return ExitGeneral
}
```

- [ ] **Step 5: Fix test — add missing `fmt` import in test file**

The test file needs `"fmt"` import for `TestGetExitCode_generic`. Add it.

- [ ] **Step 6: Run test to verify it passes**

```bash
cd /root/dev/planitai/houjin-cli && go test ./internal/errors/... -v
```

Expected: all 5 tests PASS.

- [ ] **Step 7: Create main.go stub**

Create `main.go`:

```go
package main

import "github.com/planitaicojp/houjin-cli/cmd"

func main() {
	cmd.Execute()
}
```

This won't compile yet — `cmd` package doesn't exist. That's fine; it's created in Task 3.

- [ ] **Step 8: Commit**

```bash
git add go.mod internal/errors/ main.go
git commit -m "feat: project scaffold with errors package and exit codes"
```

---

## Task 2: Model Package — Corporation Struct + Check Digit

**Files:**
- Create: `internal/model/corporation.go`, `internal/model/corporation_test.go`
- Create: `testdata/num_response.xml`

- [ ] **Step 1: Write check digit and XML parsing tests**

Create `internal/model/corporation_test.go`:

```go
package model_test

import (
	"encoding/xml"
	"os"
	"testing"

	"github.com/planitaicojp/houjin-cli/internal/model"
)

func TestValidateCorporateNumber_valid(t *testing.T) {
	// トヨタ自動車の法人番号
	if err := model.ValidateCorporateNumber("2180301018771"); err != nil {
		t.Errorf("expected valid, got: %v", err)
	}
}

func TestValidateCorporateNumber_invalidCheckDigit(t *testing.T) {
	err := model.ValidateCorporateNumber("1180301018771")
	if err == nil {
		t.Error("expected error for invalid check digit")
	}
}

func TestValidateCorporateNumber_wrongLength(t *testing.T) {
	err := model.ValidateCorporateNumber("123")
	if err == nil {
		t.Error("expected error for wrong length")
	}
}

func TestValidateCorporateNumber_nonNumeric(t *testing.T) {
	err := model.ValidateCorporateNumber("abcdefghijklm")
	if err == nil {
		t.Error("expected error for non-numeric")
	}
}

func TestParseXMLResponse(t *testing.T) {
	data, err := os.ReadFile("../../testdata/num_response.xml")
	if err != nil {
		t.Fatalf("failed to read testdata: %v", err)
	}

	var xmlResp model.XMLResponse
	if err := xml.Unmarshal(data, &xmlResp); err != nil {
		t.Fatalf("failed to parse XML: %v", err)
	}

	resp := xmlResp.ToResponse()
	if resp.Count != 1 {
		t.Errorf("expected count 1, got %d", resp.Count)
	}
	if len(resp.Corporations) != 1 {
		t.Fatalf("expected 1 corporation, got %d", len(resp.Corporations))
	}
	corp := resp.Corporations[0]
	if corp.CorporateNumber != "2180301018771" {
		t.Errorf("unexpected corporate number: %s", corp.CorporateNumber)
	}
	if corp.Name != "トヨタ自動車株式会社" {
		t.Errorf("unexpected name: %s", corp.Name)
	}
}
```

- [ ] **Step 2: Create test XML data**

Create `testdata/num_response.xml`:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<corporations>
  <lastUpdateDate>2024-01-15</lastUpdateDate>
  <count>1</count>
  <divideNumber>1</divideNumber>
  <divideSize>1</divideSize>
  <corporation>
    <sequenceNumber>1</sequenceNumber>
    <corporateNumber>2180301018771</corporateNumber>
    <process>01</process>
    <correct>0</correct>
    <updateDate>2024-01-10</updateDate>
    <changeDate>2015-10-05</changeDate>
    <name>トヨタ自動車株式会社</name>
    <nameImageId/>
    <kind>301</kind>
    <prefectureName>愛知県</prefectureName>
    <cityName>豊田市</cityName>
    <streetNumber>トヨタ町１番地</streetNumber>
    <addressImageId/>
    <prefectureCode>23</prefectureCode>
    <cityCode>211</cityCode>
    <postCode>4718571</postCode>
    <addressOutside/>
    <addressOutsideImageId/>
    <closeDate/>
    <closeCause/>
    <successorCorporateNumber/>
    <changeCause/>
    <assignmentDate>2015-10-05</assignmentDate>
    <latest>1</latest>
    <enName>TOYOTA MOTOR CORPORATION</enName>
    <enPrefectureName>Aichi</enPrefectureName>
    <enCityName>Toyota City</enCityName>
    <enAddressOutside/>
    <furigana>トヨタジドウシャ</furigana>
    <hihyoji>0</hihyoji>
  </corporation>
</corporations>
```

- [ ] **Step 3: Run test to verify it fails**

```bash
cd /root/dev/planitai/houjin-cli && go test ./internal/model/... -v
```

Expected: compilation error — model package doesn't exist.

- [ ] **Step 4: Implement model package**

Create `internal/model/corporation.go`:

```go
package model

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"unicode"
)

// XMLResponse represents the raw XML response from the API.
type XMLResponse struct {
	XMLName       xml.Name         `xml:"corporations"`
	Count         int              `xml:"count"`
	DivideNumber  int              `xml:"divideNumber"`
	DivideSize    int              `xml:"divideSize"`
	Corporations  []XMLCorporation `xml:"corporation"`
}

// XMLCorporation represents a single corporation entry in the XML response.
type XMLCorporation struct {
	CorporateNumber string `xml:"corporateNumber"`
	Name            string `xml:"name"`
	Furigana        string `xml:"furigana"`
	EnName          string `xml:"enName"`
	Kind            string `xml:"kind"`
	PrefectureName  string `xml:"prefectureName"`
	CityName        string `xml:"cityName"`
	StreetNumber    string `xml:"streetNumber"`
	PrefectureCode  string `xml:"prefectureCode"`
	CityCode        string `xml:"cityCode"`
	PostCode        string `xml:"postCode"`
	AddressOutside  string `xml:"addressOutside"`
	AssignmentDate  string `xml:"assignmentDate"`
	UpdateDate      string `xml:"updateDate"`
	ChangeDate      string `xml:"changeDate"`
	CloseDate       string `xml:"closeDate"`
	CloseCause      string `xml:"closeCause"`
	Latest          string `xml:"latest"`
	Process         string `xml:"process"`
	Correct         string `xml:"correct"`
	Hihyoji         string `xml:"hihyoji"`
}

// Corporation is the public-facing corporation data structure.
type Corporation struct {
	CorporateNumber string `json:"corporate_number"`
	Name            string `json:"name"`
	NameKana        string `json:"name_kana"`
	NameEnglish     string `json:"name_english"`
	Kind            string `json:"kind"`
	Prefecture      string `json:"prefecture"`
	City            string `json:"city"`
	Address         string `json:"address"`
	PostalCode      string `json:"postal_code"`
	AssignmentDate  string `json:"assignment_date"`
	UpdateDate      string `json:"update_date"`
	ChangeDate      string `json:"change_date"`
	CloseDate       string `json:"close_date,omitempty"`
	CloseCause      string `json:"close_cause,omitempty"`
}

// Response is the parsed API response.
type Response struct {
	Count        int           `json:"count"`
	DivideNumber int           `json:"divide_number"`
	DivideSize   int           `json:"divide_size"`
	Corporations []Corporation `json:"corporations"`
}

// ToResponse converts the XML response to the public Response type.
func (x *XMLResponse) ToResponse() *Response {
	corps := make([]Corporation, len(x.Corporations))
	for i, xc := range x.Corporations {
		corps[i] = Corporation{
			CorporateNumber: xc.CorporateNumber,
			Name:            xc.Name,
			NameKana:        xc.Furigana,
			NameEnglish:     xc.EnName,
			Kind:            xc.Kind,
			Prefecture:      xc.PrefectureName,
			City:            xc.CityName,
			Address:         xc.StreetNumber,
			PostalCode:      xc.PostCode,
			AssignmentDate:  xc.AssignmentDate,
			UpdateDate:      xc.UpdateDate,
			ChangeDate:      xc.ChangeDate,
			CloseDate:       xc.CloseDate,
			CloseCause:      xc.CloseCause,
		}
	}
	return &Response{
		Count:        x.Count,
		DivideNumber: x.DivideNumber,
		DivideSize:   x.DivideSize,
		Corporations: corps,
	}
}

// ValidateCorporateNumber validates a 13-digit corporate number including check digit.
// Check digit algorithm: https://www.houjin-bangou.nta.go.jp/shitsumon/shitsumon-07.html
func ValidateCorporateNumber(number string) error {
	if len(number) != 13 {
		return fmt.Errorf("corporate number must be 13 digits, got %d", len(number))
	}
	for _, r := range number {
		if !unicode.IsDigit(r) {
			return fmt.Errorf("corporate number must contain only digits")
		}
	}

	digits := make([]int, 13)
	for i, r := range number {
		digits[i], _ = strconv.Atoi(string(r))
	}

	// Check digit is the first digit.
	// Formula: check = 9 - (sum mod 9)
	// where sum = Σ(i=2..13) Qi × Pi
	// Pi = (i mod 2 == 0) ? 1 : 2 for positions 2..13 (1-indexed)
	sum := 0
	for i := 1; i < 13; i++ {
		p := 1
		if i%2 == 0 {
			p = 2
		}
		sum += digits[i] * p
	}
	remainder := sum % 9
	expected := 9 - remainder

	if digits[0] != expected {
		return fmt.Errorf("invalid check digit: expected %d, got %d", expected, digits[0])
	}
	return nil
}
```

- [ ] **Step 5: Run test to verify it passes**

```bash
cd /root/dev/planitai/houjin-cli && go test ./internal/model/... -v
```

Expected: all 5 tests PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/model/ testdata/
git commit -m "feat: corporation model with XML parsing and check digit validation"
```

---

## Task 3: Config Package

**Files:**
- Create: `internal/config/config.go`, `internal/config/env.go`, `internal/config/config_test.go`

- [ ] **Step 1: Write config tests**

Create `internal/config/config_test.go`:

```go
package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/planitaicojp/houjin-cli/internal/config"
)

func TestEnvOr_set(t *testing.T) {
	t.Setenv("HOUJIN_APP_ID", "test-id")
	if v := config.EnvOr("HOUJIN_APP_ID", "fallback"); v != "test-id" {
		t.Errorf("expected test-id, got %s", v)
	}
}

func TestEnvOr_fallback(t *testing.T) {
	if v := config.EnvOr("HOUJIN_NONEXISTENT", "fallback"); v != "fallback" {
		t.Errorf("expected fallback, got %s", v)
	}
}

func TestLoad_noFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOUJIN_CONFIG_DIR", dir)

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Format != "json" {
		t.Errorf("expected default format json, got %s", cfg.Format)
	}
}

func TestLoad_withFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOUJIN_CONFIG_DIR", dir)

	yamlContent := []byte("app_id: my-app-id\nformat: table\n")
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), yamlContent, 0600); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.AppID != "my-app-id" {
		t.Errorf("expected my-app-id, got %s", cfg.AppID)
	}
	if cfg.Format != "table" {
		t.Errorf("expected table, got %s", cfg.Format)
	}
}

func TestGetAppID_envOverridesFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOUJIN_CONFIG_DIR", dir)
	t.Setenv("HOUJIN_APP_ID", "env-id")

	yamlContent := []byte("app_id: file-id\n")
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), yamlContent, 0600); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	appID := config.GetAppID(cfg)
	if appID != "env-id" {
		t.Errorf("expected env-id, got %s", appID)
	}
}

func TestGetAppID_fromFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOUJIN_CONFIG_DIR", dir)
	os.Unsetenv("HOUJIN_APP_ID")

	yamlContent := []byte("app_id: file-id\n")
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), yamlContent, 0600); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	appID := config.GetAppID(cfg)
	if appID != "file-id" {
		t.Errorf("expected file-id, got %s", appID)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd /root/dev/planitai/houjin-cli && go test ./internal/config/... -v
```

Expected: compilation error.

- [ ] **Step 3: Implement env.go**

Create `internal/config/env.go`:

```go
package config

import "os"

const (
	EnvAppID     = "HOUJIN_APP_ID"
	EnvFormat    = "HOUJIN_FORMAT"
	EnvConfigDir = "HOUJIN_CONFIG_DIR"
)

// EnvOr returns the environment variable value if set, otherwise the fallback.
func EnvOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
```

- [ ] **Step 4: Implement config.go**

Create `internal/config/config.go`:

```go
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	DefaultFormat = "json"
	configFile    = "config.yaml"
)

// Config represents the application configuration.
type Config struct {
	AppID  string `yaml:"app_id"`
	Format string `yaml:"format"`
}

// DefaultConfigDir returns the config directory path.
func DefaultConfigDir() string {
	if d := os.Getenv(EnvConfigDir); d != "" {
		return d
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "houjin")
}

// Load reads the config file. Returns default config if file doesn't exist.
func Load() (*Config, error) {
	dir := DefaultConfigDir()
	path := filepath.Join(dir, configFile)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return defaultConfig(), nil
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	if cfg.Format == "" {
		cfg.Format = DefaultFormat
	}
	return &cfg, nil
}

// GetAppID returns the app ID with env override.
func GetAppID(cfg *Config) string {
	if v := os.Getenv(EnvAppID); v != "" {
		return v
	}
	return cfg.AppID
}

func defaultConfig() *Config {
	return &Config{
		Format: DefaultFormat,
	}
}
```

- [ ] **Step 5: Fetch yaml.v3 dependency**

```bash
cd /root/dev/planitai/houjin-cli && go get gopkg.in/yaml.v3
```

- [ ] **Step 6: Run test to verify it passes**

```bash
cd /root/dev/planitai/houjin-cli && go test ./internal/config/... -v
```

Expected: all 6 tests PASS.

- [ ] **Step 7: Commit**

```bash
git add internal/config/ go.mod go.sum
git commit -m "feat: config package with YAML file and env variable support"
```

---

## Task 4: Output Formatters

**Files:**
- Create: `internal/output/formatter.go`, `internal/output/json.go`, `internal/output/table.go`, `internal/output/csv.go`, `internal/output/formatter_test.go`

- [ ] **Step 1: Write formatter tests**

Create `internal/output/formatter_test.go`:

```go
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
	// default should be JSON
	var buf bytes.Buffer
	if err := f.Format(&buf, testResponse()); err != nil {
		t.Fatal(err)
	}
	var result model.Response
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("default formatter should produce valid JSON: %v", err)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd /root/dev/planitai/houjin-cli && go test ./internal/output/... -v
```

Expected: compilation error.

- [ ] **Step 3: Implement formatter.go**

Create `internal/output/formatter.go`:

```go
package output

import (
	"io"

	"github.com/planitaicojp/houjin-cli/internal/model"
)

// Formatter formats and writes a response to a writer.
type Formatter interface {
	Format(w io.Writer, resp *model.Response) error
}

// New creates a formatter for the given format name.
func New(format string) Formatter {
	switch format {
	case "table":
		return &TableFormatter{}
	case "csv":
		return &CSVFormatter{}
	default:
		return &JSONFormatter{}
	}
}
```

- [ ] **Step 4: Implement json.go**

Create `internal/output/json.go`:

```go
package output

import (
	"encoding/json"
	"io"

	"github.com/planitaicojp/houjin-cli/internal/model"
)

// JSONFormatter outputs response as indented JSON.
type JSONFormatter struct{}

func (f *JSONFormatter) Format(w io.Writer, resp *model.Response) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(resp)
}
```

- [ ] **Step 5: Implement table.go**

Create `internal/output/table.go`:

```go
package output

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/planitaicojp/houjin-cli/internal/model"
)

// TableFormatter outputs response as a fixed-width table.
type TableFormatter struct{}

func (f *TableFormatter) Format(w io.Writer, resp *model.Response) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "CORPORATE_NUMBER\tNAME\tPREFECTURE\tCITY\tADDRESS")
	for _, c := range resp.Corporations {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
			c.CorporateNumber, c.Name, c.Prefecture, c.City, c.Address)
	}
	return tw.Flush()
}
```

- [ ] **Step 6: Implement csv.go**

Create `internal/output/csv.go`:

```go
package output

import (
	"encoding/csv"
	"io"

	"github.com/planitaicojp/houjin-cli/internal/model"
)

// CSVFormatter outputs response as CSV with a header row.
type CSVFormatter struct{}

func (f *CSVFormatter) Format(w io.Writer, resp *model.Response) error {
	cw := csv.NewWriter(w)
	header := []string{
		"corporate_number", "name", "name_kana", "name_english",
		"kind", "prefecture", "city", "address", "postal_code",
		"assignment_date", "update_date", "change_date",
		"close_date", "close_cause",
	}
	if err := cw.Write(header); err != nil {
		return err
	}
	for _, c := range resp.Corporations {
		row := []string{
			c.CorporateNumber, c.Name, c.NameKana, c.NameEnglish,
			c.Kind, c.Prefecture, c.City, c.Address, c.PostalCode,
			c.AssignmentDate, c.UpdateDate, c.ChangeDate,
			c.CloseDate, c.CloseCause,
		}
		if err := cw.Write(row); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}
```

- [ ] **Step 7: Run test to verify it passes**

```bash
cd /root/dev/planitai/houjin-cli && go test ./internal/output/... -v
```

Expected: all 4 tests PASS.

- [ ] **Step 8: Commit**

```bash
git add internal/output/
git commit -m "feat: output formatters (JSON, table, CSV)"
```

---

## Task 5: API Client

**Files:**
- Create: `internal/api/client.go`, `internal/api/houjin.go`, `internal/api/client_test.go`
- Create: `testdata/name_response.xml`, `testdata/diff_response.xml`

- [ ] **Step 1: Create additional test XML data**

Create `testdata/name_response.xml`:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<corporations>
  <lastUpdateDate>2024-01-15</lastUpdateDate>
  <count>2</count>
  <divideNumber>1</divideNumber>
  <divideSize>1</divideSize>
  <corporation>
    <sequenceNumber>1</sequenceNumber>
    <corporateNumber>2180301018771</corporateNumber>
    <process>01</process>
    <correct>0</correct>
    <updateDate>2024-01-10</updateDate>
    <changeDate>2015-10-05</changeDate>
    <name>トヨタ自動車株式会社</name>
    <nameImageId/>
    <kind>301</kind>
    <prefectureName>愛知県</prefectureName>
    <cityName>豊田市</cityName>
    <streetNumber>トヨタ町１番地</streetNumber>
    <addressImageId/>
    <prefectureCode>23</prefectureCode>
    <cityCode>211</cityCode>
    <postCode>4718571</postCode>
    <addressOutside/>
    <addressOutsideImageId/>
    <closeDate/>
    <closeCause/>
    <successorCorporateNumber/>
    <changeCause/>
    <assignmentDate>2015-10-05</assignmentDate>
    <latest>1</latest>
    <enName>TOYOTA MOTOR CORPORATION</enName>
    <enPrefectureName>Aichi</enPrefectureName>
    <enCityName>Toyota City</enCityName>
    <enAddressOutside/>
    <furigana>トヨタジドウシャ</furigana>
    <hihyoji>0</hihyoji>
  </corporation>
  <corporation>
    <sequenceNumber>2</sequenceNumber>
    <corporateNumber>5180301018778</corporateNumber>
    <process>01</process>
    <correct>0</correct>
    <updateDate>2024-01-10</updateDate>
    <changeDate>2015-10-05</changeDate>
    <name>トヨタ自動車九州株式会社</name>
    <nameImageId/>
    <kind>301</kind>
    <prefectureName>福岡県</prefectureName>
    <cityName>宮若市</cityName>
    <streetNumber>上有木１番地</streetNumber>
    <addressImageId/>
    <prefectureCode>40</prefectureCode>
    <cityCode>226</cityCode>
    <postCode>8230155</postCode>
    <addressOutside/>
    <addressOutsideImageId/>
    <closeDate/>
    <closeCause/>
    <successorCorporateNumber/>
    <changeCause/>
    <assignmentDate>2015-10-05</assignmentDate>
    <latest>1</latest>
    <enName/>
    <enPrefectureName>Fukuoka</enPrefectureName>
    <enCityName>Miyawaka City</enCityName>
    <enAddressOutside/>
    <furigana>トヨタジドウシャキュウシュウ</furigana>
    <hihyoji>0</hihyoji>
  </corporation>
</corporations>
```

Create `testdata/diff_response.xml`:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<corporations>
  <lastUpdateDate>2024-01-15</lastUpdateDate>
  <count>1</count>
  <divideNumber>1</divideNumber>
  <divideSize>1</divideSize>
  <corporation>
    <sequenceNumber>1</sequenceNumber>
    <corporateNumber>2180301018771</corporateNumber>
    <process>12</process>
    <correct>0</correct>
    <updateDate>2024-01-10</updateDate>
    <changeDate>2024-01-05</changeDate>
    <name>トヨタ自動車株式会社</name>
    <nameImageId/>
    <kind>301</kind>
    <prefectureName>愛知県</prefectureName>
    <cityName>豊田市</cityName>
    <streetNumber>トヨタ町１番地</streetNumber>
    <addressImageId/>
    <prefectureCode>23</prefectureCode>
    <cityCode>211</cityCode>
    <postCode>4718571</postCode>
    <addressOutside/>
    <addressOutsideImageId/>
    <closeDate/>
    <closeCause/>
    <successorCorporateNumber/>
    <changeCause/>
    <assignmentDate>2015-10-05</assignmentDate>
    <latest>1</latest>
    <enName>TOYOTA MOTOR CORPORATION</enName>
    <enPrefectureName>Aichi</enPrefectureName>
    <enCityName>Toyota City</enCityName>
    <enAddressOutside/>
    <furigana>トヨタジドウシャ</furigana>
    <hihyoji>0</hihyoji>
  </corporation>
</corporations>
```

- [ ] **Step 2: Write API client tests**

Create `internal/api/client_test.go`:

```go
package api_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/planitaicojp/houjin-cli/internal/api"
)

func setupTestServer(t *testing.T, xmlFile string) *httptest.Server {
	t.Helper()
	data, err := os.ReadFile(xmlFile)
	if err != nil {
		t.Fatalf("failed to read test data: %v", err)
	}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify app ID is present
		if r.URL.Query().Get("id") == "" {
			http.Error(w, "missing id", http.StatusUnauthorized)
			return
		}
		// Verify type=12 (XML)
		if r.URL.Query().Get("type") != "12" {
			http.Error(w, "invalid type", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/xml")
		w.Write(data)
	}))
}

func TestGetByNumber(t *testing.T) {
	ts := setupTestServer(t, "../../testdata/num_response.xml")
	defer ts.Close()

	client := api.NewClient("test-app-id", api.WithBaseURL(ts.URL))
	resp, err := client.GetByNumber([]string{"2180301018771"}, api.GetOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Count != 1 {
		t.Errorf("expected count 1, got %d", resp.Count)
	}
	if resp.Corporations[0].Name != "トヨタ自動車株式会社" {
		t.Errorf("unexpected name: %s", resp.Corporations[0].Name)
	}
}

func TestSearchByName(t *testing.T) {
	ts := setupTestServer(t, "../../testdata/name_response.xml")
	defer ts.Close()

	client := api.NewClient("test-app-id", api.WithBaseURL(ts.URL))
	resp, err := client.SearchByName("トヨタ", api.SearchOptions{Mode: "prefix"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Count != 2 {
		t.Errorf("expected count 2, got %d", resp.Count)
	}
}

func TestGetDiff(t *testing.T) {
	ts := setupTestServer(t, "../../testdata/diff_response.xml")
	defer ts.Close()

	client := api.NewClient("test-app-id", api.WithBaseURL(ts.URL))
	resp, err := client.GetDiff("2024-01-01", "2024-01-15")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Count != 1 {
		t.Errorf("expected count 1, got %d", resp.Count)
	}
}

func TestClient_missingAppID(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	defer ts.Close()

	client := api.NewClient("", api.WithBaseURL(ts.URL))
	_, err := client.GetByNumber([]string{"2180301018771"}, api.GetOptions{})
	if err == nil {
		t.Error("expected error for missing app ID")
	}
}
```

- [ ] **Step 3: Run test to verify it fails**

```bash
cd /root/dev/planitai/houjin-cli && go test ./internal/api/... -v
```

Expected: compilation error.

- [ ] **Step 4: Implement client.go**

Create `internal/api/client.go`:

```go
package api

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	cerrors "github.com/planitaicojp/houjin-cli/internal/errors"
	"github.com/planitaicojp/houjin-cli/internal/model"
)

const (
	defaultBaseURL = "https://api.houjin-bangou.nta.go.jp/4"
	defaultTimeout = 30 * time.Second
)

// Client is the HTTP client for the 法人番号 API.
type Client struct {
	httpClient *http.Client
	baseURL    string
	appID      string
	verbose    bool
}

// Option configures the Client.
type Option func(*Client)

// WithBaseURL overrides the default base URL (for testing).
func WithBaseURL(url string) Option {
	return func(c *Client) { c.baseURL = url }
}

// WithVerbose enables verbose logging.
func WithVerbose(v bool) Option {
	return func(c *Client) { c.verbose = v }
}

// NewClient creates a new API client.
func NewClient(appID string, opts ...Option) *Client {
	c := &Client{
		httpClient: &http.Client{Timeout: defaultTimeout},
		baseURL:    defaultBaseURL,
		appID:      appID,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// fetch performs an HTTP GET and parses the XML response.
func (c *Client) fetch(endpoint string, params url.Values) (*model.Response, error) {
	params.Set("id", c.appID)
	params.Set("type", "12")

	reqURL := c.baseURL + "/" + endpoint + "?" + params.Encode()

	if c.verbose {
		fmt.Fprintf(io.Discard, "GET %s\n", reqURL) // replaced by debug logging if needed
	}

	resp, err := c.httpClient.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, &cerrors.APIError{
			StatusCode: resp.StatusCode,
			Message:    http.StatusText(resp.StatusCode),
		}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	// Handle BOM if present
	body = stripBOM(body)

	var xmlResp model.XMLResponse
	if err := xml.Unmarshal(body, &xmlResp); err != nil {
		return nil, fmt.Errorf("parsing XML: %w", err)
	}

	return xmlResp.ToResponse(), nil
}

// stripBOM removes a UTF-8 BOM prefix if present.
func stripBOM(b []byte) []byte {
	if len(b) >= 3 && b[0] == 0xEF && b[1] == 0xBB && b[2] == 0xBF {
		return b[3:]
	}
	return b
}

// joinNumbers joins corporate numbers with semicolons for the API.
func joinNumbers(numbers []string) string {
	return strings.Join(numbers, ";")
}
```

- [ ] **Step 5: Implement houjin.go**

Create `internal/api/houjin.go`:

```go
package api

import (
	"net/url"

	"github.com/planitaicojp/houjin-cli/internal/model"
)

// GetOptions configures the GetByNumber request.
type GetOptions struct {
	History bool
	Close   bool
}

// SearchOptions configures the SearchByName request.
type SearchOptions struct {
	Mode string // "prefix" or "partial"
	Pref string // prefecture code
	City string // city code
	Close bool
}

// GetByNumber fetches corporation info by corporate number(s).
func (c *Client) GetByNumber(numbers []string, opts GetOptions) (*model.Response, error) {
	params := url.Values{}
	params.Set("number", joinNumbers(numbers))
	if opts.History {
		params.Set("history", "1")
	} else {
		params.Set("history", "0")
	}
	if opts.Close {
		params.Set("close", "1")
	} else {
		params.Set("close", "0")
	}
	return c.fetch("num", params)
}

// SearchByName searches corporations by name.
func (c *Client) SearchByName(name string, opts SearchOptions) (*model.Response, error) {
	params := url.Values{}
	params.Set("name", name)

	mode := "1" // prefix
	if opts.Mode == "partial" {
		mode = "2"
	}
	params.Set("mode", mode)

	if opts.Pref != "" {
		params.Set("address", opts.Pref+opts.City)
	}
	if opts.Close {
		params.Set("close", "1")
	} else {
		params.Set("close", "0")
	}
	params.Set("target", "1")

	return c.fetch("name", params)
}

// GetDiff fetches corporations updated within the specified date range.
func (c *Client) GetDiff(from, to string) (*model.Response, error) {
	params := url.Values{}
	params.Set("from", from)
	params.Set("to", to)
	return c.fetch("diff", params)
}
```

- [ ] **Step 6: Run test to verify it passes**

```bash
cd /root/dev/planitai/houjin-cli && go test ./internal/api/... -v
```

Expected: all 4 tests PASS.

- [ ] **Step 7: Commit**

```bash
git add internal/api/ testdata/name_response.xml testdata/diff_response.xml
git commit -m "feat: API client with XML parsing for num, name, diff endpoints"
```

---

## Task 6: Root Command + Version + Completion

**Files:**
- Create: `cmd/root.go`, `cmd/version.go`, `cmd/completion.go`

- [ ] **Step 1: Fetch cobra dependency**

```bash
cd /root/dev/planitai/houjin-cli && go get github.com/spf13/cobra
```

- [ ] **Step 2: Implement root.go**

Create `cmd/root.go`:

```go
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/planitaicojp/houjin-cli/internal/config"
	cerrors "github.com/planitaicojp/houjin-cli/internal/errors"
)

var (
	version = "dev"

	flagFormat  string
	flagVerbose bool
	flagConfig  string
)

var rootCmd = &cobra.Command{
	Use:           "houjin",
	Short:         "法人番号システム Web-API CLI",
	Long:          "国税庁法人番号公表サイトのWeb-APIを操作するCLIツール",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&flagFormat, "format", "", "出力形式: json, table, csv (デフォルト: json)")
	rootCmd.PersistentFlags().BoolVar(&flagVerbose, "verbose", false, "詳細ログ出力")
	rootCmd.PersistentFlags().StringVar(&flagConfig, "config", "", "設定ファイルパス")

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(completionCmd)
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(cerrors.GetExitCode(err))
	}
}

// getFormat returns the output format from flag > env > config > default.
func getFormat() string {
	if flagFormat != "" {
		return flagFormat
	}
	if f := config.EnvOr(config.EnvFormat, ""); f != "" {
		return f
	}
	cfg, err := loadConfig()
	if err != nil {
		return config.DefaultFormat
	}
	return cfg.Format
}

// loadConfig loads the config file, respecting --config flag.
func loadConfig() (*config.Config, error) {
	if flagConfig != "" {
		os.Setenv(config.EnvConfigDir, flagConfig)
	}
	return config.Load()
}

// getAppID returns the app ID or exits with config error.
func getAppID() (string, error) {
	cfg, err := loadConfig()
	if err != nil {
		return "", err
	}
	appID := config.GetAppID(cfg)
	if appID == "" {
		return "", &cerrors.ConfigError{
			Message: "アプリケーションIDが設定されていません。\n" +
				"  環境変数: export HOUJIN_APP_ID=your-id\n" +
				"  設定ファイル: ~/.config/houjin/config.yaml に app_id を設定",
		}
	}
	return appID, nil
}
```

- [ ] **Step 3: Implement version.go**

Create `cmd/version.go`:

```go
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "バージョン情報を表示",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("houjin version %s\n", version)
	},
}
```

- [ ] **Step 4: Implement completion.go**

Create `cmd/completion.go`:

```go
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish]",
	Short: "シェル補完スクリプトを生成",
	Long: `シェル補完スクリプトを生成します。

  bash:  source <(houjin completion bash)
  zsh:   houjin completion zsh > "${fpath[1]}/_houjin"
  fish:  houjin completion fish | source`,
	Args:      cobra.ExactArgs(1),
	ValidArgs: []string{"bash", "zsh", "fish"},
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return rootCmd.GenBashCompletion(os.Stdout)
		case "zsh":
			return rootCmd.GenZshCompletion(os.Stdout)
		case "fish":
			return rootCmd.GenFishCompletion(os.Stdout, true)
		default:
			return cmd.Usage()
		}
	},
}
```

- [ ] **Step 5: Verify build compiles**

```bash
cd /root/dev/planitai/houjin-cli && go build ./...
```

Expected: successful compilation.

- [ ] **Step 6: Commit**

```bash
git add cmd/root.go cmd/version.go cmd/completion.go main.go go.mod go.sum
git commit -m "feat: root command with version, completion, and global flags"
```

---

## Task 7: Get Subcommand

**Files:**
- Create: `cmd/get.go`

- [ ] **Step 1: Implement get.go**

Create `cmd/get.go`:

```go
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/planitaicojp/houjin-cli/internal/api"
	"github.com/planitaicojp/houjin-cli/internal/errors"
	"github.com/planitaicojp/houjin-cli/internal/model"
	"github.com/planitaicojp/houjin-cli/internal/output"
)

var (
	getHistory bool
	getClose   bool
)

func init() {
	getCmd.Flags().BoolVar(&getHistory, "history", false, "履歴情報を含める")
	getCmd.Flags().BoolVar(&getClose, "close", false, "閉鎖法人を含める")
	rootCmd.AddCommand(getCmd)
}

var getCmd = &cobra.Command{
	Use:   "get <法人番号> [法人番号...]",
	Short: "法人番号を指定して法人情報を取得",
	Long:  "13桁の法人番号を指定して法人情報を取得します。複数番号を空白区切りで指定可能。",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Validate all corporate numbers first
		for _, num := range args {
			if err := model.ValidateCorporateNumber(num); err != nil {
				return &errors.ValidationError{
					Field:   "corporate_number",
					Message: fmt.Sprintf("%s: %v", num, err),
				}
			}
		}

		appID, err := getAppID()
		if err != nil {
			return err
		}

		client := api.NewClient(appID, api.WithVerbose(flagVerbose))
		resp, err := client.GetByNumber(args, api.GetOptions{
			History: getHistory,
			Close:   getClose,
		})
		if err != nil {
			return err
		}

		formatter := output.New(getFormat())
		return formatter.Format(os.Stdout, resp)
	},
}
```

- [ ] **Step 2: Verify build compiles**

```bash
cd /root/dev/planitai/houjin-cli && go build ./...
```

Expected: successful compilation.

- [ ] **Step 3: Commit**

```bash
git add cmd/get.go
git commit -m "feat: get subcommand for corporate number lookup"
```

---

## Task 8: Search Subcommand

**Files:**
- Create: `cmd/search.go`

- [ ] **Step 1: Implement search.go**

Create `cmd/search.go`:

```go
package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/planitaicojp/houjin-cli/internal/api"
	"github.com/planitaicojp/houjin-cli/internal/output"
)

var (
	searchMode  string
	searchPref  string
	searchCity  string
	searchClose bool
)

func init() {
	searchCmd.Flags().StringVar(&searchMode, "mode", "prefix", "検索モード: prefix(前方一致), partial(部分一致)")
	searchCmd.Flags().StringVar(&searchPref, "pref", "", "都道府県コード (01-47, 99=海外)")
	searchCmd.Flags().StringVar(&searchCity, "city", "", "市区町村コード")
	searchCmd.Flags().BoolVar(&searchClose, "close", false, "閉鎖法人を含める")
	rootCmd.AddCommand(searchCmd)
}

var searchCmd = &cobra.Command{
	Use:   "search <法人名>",
	Short: "法人名で検索",
	Long:  "法人名を指定して法人情報を検索します。前方一致（デフォルト）または部分一致で検索可能。",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		appID, err := getAppID()
		if err != nil {
			return err
		}

		client := api.NewClient(appID, api.WithVerbose(flagVerbose))
		resp, err := client.SearchByName(args[0], api.SearchOptions{
			Mode:  searchMode,
			Pref:  searchPref,
			City:  searchCity,
			Close: searchClose,
		})
		if err != nil {
			return err
		}

		formatter := output.New(getFormat())
		return formatter.Format(os.Stdout, resp)
	},
}
```

- [ ] **Step 2: Verify build compiles**

```bash
cd /root/dev/planitai/houjin-cli && go build ./...
```

Expected: successful compilation.

- [ ] **Step 3: Commit**

```bash
git add cmd/search.go
git commit -m "feat: search subcommand for corporate name lookup"
```

---

## Task 9: Diff Subcommand

**Files:**
- Create: `cmd/diff.go`

- [ ] **Step 1: Implement diff.go**

Create `cmd/diff.go`:

```go
package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/planitaicojp/houjin-cli/internal/api"
	"github.com/planitaicojp/houjin-cli/internal/output"
)

var (
	diffFrom string
	diffTo   string
)

func init() {
	diffCmd.Flags().StringVar(&diffFrom, "from", "", "開始日 (YYYY-MM-DD) (必須)")
	diffCmd.Flags().StringVar(&diffTo, "to", "", "終了日 (YYYY-MM-DD) (必須)")
	diffCmd.MarkFlagRequired("from")
	diffCmd.MarkFlagRequired("to")
	rootCmd.AddCommand(diffCmd)
}

var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "指定期間内の更新法人一覧を取得",
	Long:  "指定した期間内に更新された法人の一覧を取得します。",
	RunE: func(cmd *cobra.Command, args []string) error {
		appID, err := getAppID()
		if err != nil {
			return err
		}

		client := api.NewClient(appID, api.WithVerbose(flagVerbose))
		resp, err := client.GetDiff(diffFrom, diffTo)
		if err != nil {
			return err
		}

		formatter := output.New(getFormat())
		return formatter.Format(os.Stdout, resp)
	},
}
```

- [ ] **Step 2: Verify build compiles**

```bash
cd /root/dev/planitai/houjin-cli && go build ./...
```

Expected: successful compilation.

- [ ] **Step 3: Commit**

```bash
git add cmd/diff.go
git commit -m "feat: diff subcommand for update history lookup"
```

---

## Task 10: Full Build + All Tests Green

**Files:**
- No new files

- [ ] **Step 1: Run all tests**

```bash
cd /root/dev/planitai/houjin-cli && go test ./... -v
```

Expected: all tests PASS across errors, model, config, output, api packages.

- [ ] **Step 2: Verify full build**

```bash
cd /root/dev/planitai/houjin-cli && go build -o houjin .
```

Expected: `houjin` binary created.

- [ ] **Step 3: Smoke test CLI**

```bash
./houjin --help
./houjin version
./houjin get --help
./houjin search --help
./houjin diff --help
```

Expected: all help texts display correctly with Japanese descriptions.

- [ ] **Step 4: Verify error for missing app ID**

```bash
unset HOUJIN_APP_ID && ./houjin get 2180301018771
```

Expected: config error message with setup instructions.

- [ ] **Step 5: Clean up and final commit**

```bash
cd /root/dev/planitai/houjin-cli && go mod tidy
git add go.mod go.sum
git commit -m "chore: go mod tidy"
```
