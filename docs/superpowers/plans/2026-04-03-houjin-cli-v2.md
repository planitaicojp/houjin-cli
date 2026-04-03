# houjin-cli v0.2.0 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Release v0.1.0, then add paging, filtering, address params, and batch input features.

**Architecture:** Each feature extends the existing cmd → api → model pipeline. New API parameters are added to option structs, new CLI flags map to those options, and the API client passes them as query parameters. Batch input reads files and feeds numbers into existing GetByNumber flow.

**Tech Stack:** Go, spf13/cobra, net/http, encoding/xml

---

### Important API Notes

- `params.Set("type", "12")` in `client.go:59` is the **response format** parameter (12 = XML+Unicode). This is NOT the corporation type filter.
- The corporation type filter in the `/4/name` endpoint is the `kind` query parameter (01-04).
- The change reason filter in the `/4/diff` endpoint is also the `kind` query parameter (01-99).
- The CLI uses `--type` flag for search (maps to API `kind`) and `--kind` flag for diff (maps to API `kind`) — different flag names because they mean different things.

---

### Task 0: Fix goreleaser config and release v0.1.0

**Files:**
- Modify: `.goreleaser.yaml`
- Modify: `.github/workflows/release.yml`
- Modify: `README.md:26-27` (scoop bucket URL)

- [ ] **Step 1: Update `.goreleaser.yaml` — add token fields, fix scoop repo name**

Add `token` to brews and scoops sections, change scoop repo name from `scoop-bucket` to `bucket` to match freee-cli convention.

```yaml
brews:
  - repository:
      owner: planitaicojp
      name: homebrew-tap
      token: "{{ .Env.HOMEBREW_TAP_TOKEN }}"
    homepage: https://github.com/planitaicojp/houjin-cli
    description: "法人番号システム Web-API CLI tool"
    license: MIT
    install: |
      bin.install "houjin"

scoops:
  - repository:
      owner: planitaicojp
      name: bucket
      token: "{{ .Env.HOMEBREW_TAP_TOKEN }}"
    homepage: https://github.com/planitaicojp/houjin-cli
    description: "法人番号システム Web-API CLI tool"
    license: MIT
```

- [ ] **Step 2: Update `.github/workflows/release.yml` — add HOMEBREW_TAP_TOKEN env**

```yaml
      - uses: goreleaser/goreleaser-action@v6
        with:
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          HOMEBREW_TAP_TOKEN: ${{ secrets.HOMEBREW_TAP_TOKEN }}
```

- [ ] **Step 3: Update `README.md` scoop bucket URL**

Change:
```
scoop bucket add planitaicojp https://github.com/planitaicojp/scoop-bucket
```
To:
```
scoop bucket add planitaicojp https://github.com/planitaicojp/bucket
```

- [ ] **Step 4: Run tests to verify no regressions**

Run: `cd /root/dev/planitai/houjin-cli && go test ./... -v`
Expected: All 21 tests pass.

- [ ] **Step 5: Commit**

```bash
git add .goreleaser.yaml .github/workflows/release.yml README.md
git commit -m "fix: goreleaser token config and scoop bucket name"
```

- [ ] **Step 6: Create PR and merge**

Create a PR from feature branch to main. After merge, tag and push:

```bash
git tag v0.1.0
git push origin v0.1.0
```

- [ ] **Step 7: Verify release**

Check GitHub Actions release workflow completes. Verify:
- GitHub Releases page has binaries for linux/darwin/windows × amd64/arm64
- checksums.txt is present
- `planitaicojp/homebrew-tap` has houjin formula
- `planitaicojp/bucket` has houjin manifest

---

### Task 1: Add paging support (`--page` and `--all` flags)

**Files:**
- Modify: `internal/api/houjin.go` — add Divide field to SearchOptions/DiffOptions, add FetchAll methods
- Modify: `internal/api/client.go` — add fetchRaw for internal reuse
- Create: `internal/api/houjin_test.go` — paging-specific tests
- Modify: `cmd/search.go` — add `--page` and `--all` flags
- Modify: `cmd/diff.go` — add `--page` and `--all` flags, add DiffOptions struct usage

**Testdata:**
- Create: `testdata/name_response_page1.xml` — page 1 of 2
- Create: `testdata/name_response_page2.xml` — page 2 of 2

- [ ] **Step 1: Create paging test data**

Create `testdata/name_response_page1.xml`:
```xml
<?xml version="1.0" encoding="UTF-8"?>
<corporations>
  <lastUpdateDate>2024-01-15</lastUpdateDate>
  <count>2</count>
  <divideNumber>1</divideNumber>
  <divideSize>2</divideSize>
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

Create `testdata/name_response_page2.xml`:
```xml
<?xml version="1.0" encoding="UTF-8"?>
<corporations>
  <lastUpdateDate>2024-01-15</lastUpdateDate>
  <count>2</count>
  <divideNumber>2</divideNumber>
  <divideSize>2</divideSize>
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

- [ ] **Step 2: Write failing test for paging in SearchByName**

Add to `internal/api/client_test.go`:

```go
func TestSearchByName_withPage(t *testing.T) {
	ts := setupTestServer(t, "../../testdata/name_response_page2.xml")
	defer ts.Close()

	client := api.NewClient("test-app-id", api.WithBaseURL(ts.URL))
	resp, err := client.SearchByName("トヨタ", api.SearchOptions{
		Mode:   "prefix",
		Divide: 2,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.DivideNumber != 2 {
		t.Errorf("expected divide_number 2, got %d", resp.DivideNumber)
	}
}
```

Run: `cd /root/dev/planitai/houjin-cli && go test ./internal/api/ -run TestSearchByName_withPage -v`
Expected: FAIL — `api.SearchOptions` has no field `Divide`.

- [ ] **Step 3: Add Divide field to SearchOptions and DiffOptions, wire query param**

In `internal/api/houjin.go`:

Add `Divide int` to `SearchOptions`:
```go
type SearchOptions struct {
	Mode   string
	Pref   string
	City   string
	Close  bool
	Kind   string // corporation type filter (01-04)
	Divide int    // page number (0 = default/first page)
}
```

Create `DiffOptions` struct:
```go
// DiffOptions configures the GetDiff request.
type DiffOptions struct {
	Kind   string // change reason filter (01-99)
	Divide int    // page number (0 = default/first page)
}
```

In `SearchByName`, add after existing params:
```go
if opts.Divide > 0 {
	params.Set("divide", strconv.Itoa(opts.Divide))
}
```

Update `GetDiff` signature to accept `DiffOptions`:
```go
func (c *Client) GetDiff(from, to string, opts DiffOptions) (*model.Response, error) {
	params := url.Values{}
	params.Set("from", from)
	params.Set("to", to)
	if opts.Divide > 0 {
		params.Set("divide", strconv.Itoa(opts.Divide))
	}
	return c.fetch("diff", params)
}
```

Add `import "strconv"` to the import block.

- [ ] **Step 4: Update GetDiff callers**

In `cmd/diff.go`, update the call:
```go
resp, err := client.GetDiff(diffFrom, diffTo, api.DiffOptions{})
```

In `internal/api/client_test.go`, update `TestGetDiff`:
```go
resp, err := client.GetDiff("2024-01-01", "2024-01-15", api.DiffOptions{})
```

- [ ] **Step 5: Run tests to verify paging param works**

Run: `cd /root/dev/planitai/houjin-cli && go test ./... -v`
Expected: All tests pass including `TestSearchByName_withPage`.

- [ ] **Step 6: Write failing test for SearchAllPages**

Add to `internal/api/client_test.go`:

```go
func setupPagingServer(t *testing.T) *httptest.Server {
	t.Helper()
	page1, err := os.ReadFile("../../testdata/name_response_page1.xml")
	if err != nil {
		t.Fatalf("failed to read page1: %v", err)
	}
	page2, err := os.ReadFile("../../testdata/name_response_page2.xml")
	if err != nil {
		t.Fatalf("failed to read page2: %v", err)
	}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("id") == "" {
			http.Error(w, "missing id", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/xml")
		divide := r.URL.Query().Get("divide")
		if divide == "2" {
			w.Write(page2)
		} else {
			w.Write(page1)
		}
	}))
}

func TestSearchAllPages(t *testing.T) {
	ts := setupPagingServer(t)
	defer ts.Close()

	client := api.NewClient("test-app-id", api.WithBaseURL(ts.URL))
	resp, err := client.SearchAllPages("トヨタ", api.SearchOptions{Mode: "prefix"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Count != 2 {
		t.Errorf("expected count 2, got %d", resp.Count)
	}
	if len(resp.Corporations) != 2 {
		t.Errorf("expected 2 corporations, got %d", len(resp.Corporations))
	}
}
```

Run: `cd /root/dev/planitai/houjin-cli && go test ./internal/api/ -run TestSearchAllPages -v`
Expected: FAIL — `SearchAllPages` method not defined.

- [ ] **Step 7: Implement SearchAllPages and DiffAllPages**

In `internal/api/houjin.go`, add:

```go
// SearchAllPages fetches all pages of search results and merges them.
func (c *Client) SearchAllPages(name string, opts SearchOptions) (*model.Response, error) {
	opts.Divide = 0
	first, err := c.SearchByName(name, opts)
	if err != nil {
		return nil, err
	}
	if first.DivideSize <= 1 {
		return first, nil
	}

	all := first.Corporations
	for page := 2; page <= first.DivideSize; page++ {
		opts.Divide = page
		resp, err := c.SearchByName(name, opts)
		if err != nil {
			return nil, fmt.Errorf("fetching page %d: %w", page, err)
		}
		all = append(all, resp.Corporations...)
	}
	first.Corporations = all
	return first, nil
}

// DiffAllPages fetches all pages of diff results and merges them.
func (c *Client) DiffAllPages(from, to string, opts DiffOptions) (*model.Response, error) {
	opts.Divide = 0
	first, err := c.GetDiff(from, to, opts)
	if err != nil {
		return nil, err
	}
	if first.DivideSize <= 1 {
		return first, nil
	}

	all := first.Corporations
	for page := 2; page <= first.DivideSize; page++ {
		opts.Divide = page
		resp, err := c.GetDiff(from, to, opts)
		if err != nil {
			return nil, fmt.Errorf("fetching page %d: %w", page, err)
		}
		all = append(all, resp.Corporations...)
	}
	first.Corporations = all
	return first, nil
}
```

Add `"fmt"` to the import block in `houjin.go`.

- [ ] **Step 8: Run tests**

Run: `cd /root/dev/planitai/houjin-cli && go test ./... -v`
Expected: All tests pass.

- [ ] **Step 9: Add `--page` and `--all` flags to search command**

In `cmd/search.go`, add variables:

```go
var (
	searchMode  string
	searchPref  string
	searchCity  string
	searchClose bool
	searchPage  int
	searchAll   bool
)
```

In `init()`, add:
```go
searchCmd.Flags().IntVar(&searchPage, "page", 0, "ページ番号を指定 (分割番号)")
searchCmd.Flags().BoolVar(&searchAll, "all", false, "全ページを自動取得")
```

Update `RunE`:
```go
RunE: func(cmd *cobra.Command, args []string) error {
	appID, err := getAppID()
	if err != nil {
		return err
	}

	client := api.NewClient(appID, api.WithVerbose(flagVerbose))
	opts := api.SearchOptions{
		Mode:   searchMode,
		Pref:   searchPref,
		City:   searchCity,
		Close:  searchClose,
		Divide: searchPage,
	}

	var resp *model.Response
	if searchAll {
		resp, err = client.SearchAllPages(args[0], opts)
	} else {
		resp, err = client.SearchByName(args[0], opts)
	}
	if err != nil {
		return err
	}

	formatter := output.New(getFormat())
	return formatter.Format(os.Stdout, resp)
},
```

Add `"github.com/planitaicojp/houjin-cli/internal/model"` to imports.

- [ ] **Step 10: Add `--page` and `--all` flags to diff command**

In `cmd/diff.go`, add variables:

```go
var (
	diffFrom string
	diffTo   string
	diffPage int
	diffAll  bool
)
```

In `init()`, add:
```go
diffCmd.Flags().IntVar(&diffPage, "page", 0, "ページ番号を指定 (分割番号)")
diffCmd.Flags().BoolVar(&diffAll, "all", false, "全ページを自動取得")
```

Update `RunE`:
```go
RunE: func(cmd *cobra.Command, args []string) error {
	appID, err := getAppID()
	if err != nil {
		return err
	}

	client := api.NewClient(appID, api.WithVerbose(flagVerbose))
	opts := api.DiffOptions{
		Divide: diffPage,
	}

	var resp *model.Response
	if diffAll {
		resp, err = client.DiffAllPages(diffFrom, diffTo, opts)
	} else {
		resp, err = client.GetDiff(diffFrom, diffTo, opts)
	}
	if err != nil {
		return err
	}

	formatter := output.New(getFormat())
	return formatter.Format(os.Stdout, resp)
},
```

Add `"github.com/planitaicojp/houjin-cli/internal/model"` to imports.

- [ ] **Step 11: Run all tests**

Run: `cd /root/dev/planitai/houjin-cli && go test ./... -v`
Expected: All tests pass.

- [ ] **Step 12: Commit**

```bash
git add internal/api/houjin.go internal/api/client_test.go cmd/search.go cmd/diff.go testdata/name_response_page1.xml testdata/name_response_page2.xml
git commit -m "feat: add paging support (--page, --all) for search and diff"
```

---

### Task 2: Add corporation type filter (`--type` flag for search)

**Files:**
- Modify: `internal/api/houjin.go` — add Kind to SearchOptions, wire to `kind` query param
- Modify: `internal/api/client_test.go` — test for type filter
- Modify: `cmd/search.go` — add `--type` flag

Note: The CLI `--type` flag maps to the API `kind` query parameter in `/4/name`. The `Kind` field was already added to `SearchOptions` in Task 1 Step 3.

- [ ] **Step 1: Write failing test for type filter**

Add to `internal/api/client_test.go`:

```go
func TestSearchByName_withTypeFilter(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		kind := r.URL.Query().Get("kind")
		if kind != "03" {
			t.Errorf("expected kind=03, got %s", kind)
		}
		data, _ := os.ReadFile("../../testdata/name_response.xml")
		w.Header().Set("Content-Type", "application/xml")
		w.Write(data)
	}))
	defer ts.Close()

	client := api.NewClient("test-app-id", api.WithBaseURL(ts.URL))
	_, err := client.SearchByName("トヨタ", api.SearchOptions{
		Mode: "prefix",
		Kind: "03",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
```

Run: `cd /root/dev/planitai/houjin-cli && go test ./internal/api/ -run TestSearchByName_withTypeFilter -v`
Expected: Test passes but `kind` query param is not sent (test assertion inside handler will fail via t.Errorf).

- [ ] **Step 2: Wire Kind to query parameter in SearchByName**

In `internal/api/houjin.go`, in `SearchByName`, add after the `close` param:

```go
if opts.Kind != "" {
	params.Set("kind", opts.Kind)
}
```

- [ ] **Step 3: Run test to verify**

Run: `cd /root/dev/planitai/houjin-cli && go test ./internal/api/ -run TestSearchByName_withTypeFilter -v`
Expected: PASS

- [ ] **Step 4: Add `--type` flag to search command**

In `cmd/search.go`, add variable:

```go
var (
	searchMode  string
	searchPref  string
	searchCity  string
	searchClose bool
	searchPage  int
	searchAll   bool
	searchType  string
)
```

In `init()`, add:
```go
searchCmd.Flags().StringVar(&searchType, "type", "", "法人種別フィルタ (01:国の機関, 02:地方公共団体, 03:設立登記法人, 04:その他)")
```

In `RunE`, add `Kind: searchType` to the opts:
```go
opts := api.SearchOptions{
	Mode:   searchMode,
	Pref:   searchPref,
	City:   searchCity,
	Close:  searchClose,
	Kind:   searchType,
	Divide: searchPage,
}
```

- [ ] **Step 5: Run all tests**

Run: `cd /root/dev/planitai/houjin-cli && go test ./... -v`
Expected: All tests pass.

- [ ] **Step 6: Commit**

```bash
git add internal/api/houjin.go internal/api/client_test.go cmd/search.go
git commit -m "feat: add corporation type filter (--type) for search"
```

---

### Task 3: Add change reason filter (`--kind` flag for diff)

**Files:**
- Modify: `internal/api/houjin.go` — wire Kind in DiffOptions to `kind` query param
- Modify: `internal/api/client_test.go` — test for kind filter
- Modify: `cmd/diff.go` — add `--kind` flag

Note: `DiffOptions.Kind` was already added in Task 1 Step 3.

- [ ] **Step 1: Write failing test for kind filter in diff**

Add to `internal/api/client_test.go`:

```go
func TestGetDiff_withKindFilter(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		kind := r.URL.Query().Get("kind")
		if kind != "01" {
			t.Errorf("expected kind=01, got %s", kind)
		}
		data, _ := os.ReadFile("../../testdata/diff_response.xml")
		w.Header().Set("Content-Type", "application/xml")
		w.Write(data)
	}))
	defer ts.Close()

	client := api.NewClient("test-app-id", api.WithBaseURL(ts.URL))
	_, err := client.GetDiff("2024-01-01", "2024-01-15", api.DiffOptions{Kind: "01"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
```

Run: `cd /root/dev/planitai/houjin-cli && go test ./internal/api/ -run TestGetDiff_withKindFilter -v`
Expected: Test passes but `kind` param not sent (t.Errorf in handler fires).

- [ ] **Step 2: Wire Kind to query parameter in GetDiff**

In `internal/api/houjin.go`, in `GetDiff`, add after the `divide` param:

```go
if opts.Kind != "" {
	params.Set("kind", opts.Kind)
}
```

- [ ] **Step 3: Run test to verify**

Run: `cd /root/dev/planitai/houjin-cli && go test ./internal/api/ -run TestGetDiff_withKindFilter -v`
Expected: PASS

- [ ] **Step 4: Add `--kind` flag to diff command**

In `cmd/diff.go`, add variable:

```go
var (
	diffFrom string
	diffTo   string
	diffPage int
	diffAll  bool
	diffKind string
)
```

In `init()`, add:
```go
diffCmd.Flags().StringVar(&diffKind, "kind", "", "変更事由フィルタ (01:新規, 02:商号変更, 03:国内所在地変更, 04:国外所在地変更, 11:閉鎖, 12:復活, 13:吸収合併, 14:合併無効, 15:抹消, 99:削除)")
```

In `RunE`, add `Kind: diffKind` to opts:
```go
opts := api.DiffOptions{
	Kind:   diffKind,
	Divide: diffPage,
}
```

- [ ] **Step 5: Run all tests**

Run: `cd /root/dev/planitai/houjin-cli && go test ./... -v`
Expected: All tests pass.

- [ ] **Step 6: Commit**

```bash
git add internal/api/houjin.go internal/api/client_test.go cmd/diff.go
git commit -m "feat: add change reason filter (--kind) for diff"
```

---

### Task 4: Verify address parameter wiring (`--city` with `--pref`)

**Files:**
- Modify: `internal/api/client_test.go` — test for address param combination
- Possibly modify: `internal/api/houjin.go` (if city param not properly wired)

Note: Looking at current `SearchByName` in `houjin.go:52-53`, the address is built as `opts.Pref + opts.City` and set as a single `address` param. This means `--city` already works when combined with `--pref`. We just need to verify with a test and ensure `--city` without `--pref` is handled.

- [ ] **Step 1: Write test for address parameter combination**

Add to `internal/api/client_test.go`:

```go
func TestSearchByName_withAddress(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		addr := r.URL.Query().Get("address")
		if addr != "13101" {
			t.Errorf("expected address=13101, got %s", addr)
		}
		data, _ := os.ReadFile("../../testdata/name_response.xml")
		w.Header().Set("Content-Type", "application/xml")
		w.Write(data)
	}))
	defer ts.Close()

	client := api.NewClient("test-app-id", api.WithBaseURL(ts.URL))
	_, err := client.SearchByName("株式会社", api.SearchOptions{
		Mode: "prefix",
		Pref: "13",
		City: "101",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
```

Run: `cd /root/dev/planitai/houjin-cli && go test ./internal/api/ -run TestSearchByName_withAddress -v`
Expected: PASS (the current code already concatenates pref+city).

- [ ] **Step 2: Run all tests**

Run: `cd /root/dev/planitai/houjin-cli && go test ./... -v`
Expected: All tests pass.

- [ ] **Step 3: Commit**

```bash
git add internal/api/client_test.go
git commit -m "test: verify address parameter wiring (--pref + --city)"
```

---

### Task 5: Add batch input (`--file` flag for get)

**Files:**
- Create: `internal/batch/reader.go` — file reading logic
- Create: `internal/batch/reader_test.go` — tests
- Create: `testdata/batch_numbers.txt` — test input file
- Modify: `cmd/get.go` — add `--file` flag, integrate batch reader

- [ ] **Step 1: Create test input file**

Create `testdata/batch_numbers.txt`:
```
# 法人番号リスト
1180301018771

5180301018778
```

- [ ] **Step 2: Write failing test for batch reader**

Create `internal/batch/reader_test.go`:

```go
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
```

Run: `cd /root/dev/planitai/houjin-cli && go test ./internal/batch/ -v`
Expected: FAIL — package does not exist.

- [ ] **Step 3: Implement batch reader**

Create `internal/batch/reader.go`:

```go
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
```

- [ ] **Step 4: Run batch reader tests**

Run: `cd /root/dev/planitai/houjin-cli && go test ./internal/batch/ -v`
Expected: All 3 tests pass.

- [ ] **Step 5: Add `--file` flag to get command**

In `cmd/get.go`, add variable:

```go
var (
	getHistory bool
	getClose   bool
	getFile    string
)
```

In `init()`, add:
```go
getCmd.Flags().StringVar(&getFile, "file", "", "法人番号リストファイル (- で標準入力)")
```

Change `Args` from `cobra.MinimumNArgs(1)` to `cobra.ArbitraryArgs` (since `--file` mode needs 0 args).

Update `RunE`:

```go
RunE: func(cmd *cobra.Command, args []string) error {
	var numbers []string

	if getFile != "" {
		var r io.Reader
		if getFile == "-" {
			r = os.Stdin
		} else {
			f, err := os.Open(getFile)
			if err != nil {
				return fmt.Errorf("opening file: %w", err)
			}
			defer f.Close()
			r = f
		}
		var err error
		numbers, err = batch.ReadNumbers(r)
		if err != nil {
			return fmt.Errorf("reading numbers: %w", err)
		}
		if len(numbers) == 0 {
			return &cerrors.ValidationError{
				Field:   "file",
				Message: "ファイルに法人番号が含まれていません",
			}
		}
	} else {
		if len(args) == 0 {
			return &cerrors.ValidationError{
				Field:   "args",
				Message: "法人番号を指定するか、--file でファイルを指定してください",
			}
		}
		numbers = args
	}

	for _, num := range numbers {
		if err := model.ValidateCorporateNumber(num); err != nil {
			return &cerrors.ValidationError{
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
	resp, err := client.GetByNumber(numbers, api.GetOptions{
		History: getHistory,
		Close:   getClose,
	})
	if err != nil {
		return err
	}

	formatter := output.New(getFormat())
	return formatter.Format(os.Stdout, resp)
},
```

Add imports:
```go
import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/planitaicojp/houjin-cli/internal/api"
	"github.com/planitaicojp/houjin-cli/internal/batch"
	cerrors "github.com/planitaicojp/houjin-cli/internal/errors"
	"github.com/planitaicojp/houjin-cli/internal/model"
	"github.com/planitaicojp/houjin-cli/internal/output"
)
```

- [ ] **Step 6: Run all tests**

Run: `cd /root/dev/planitai/houjin-cli && go test ./... -v`
Expected: All tests pass.

- [ ] **Step 7: Commit**

```bash
git add internal/batch/reader.go internal/batch/reader_test.go testdata/batch_numbers.txt cmd/get.go
git commit -m "feat: add batch input (--file) for get command"
```

---

### Task 6: Update README with new features and App ID guide

**Files:**
- Modify: `README.md`

- [ ] **Step 1: Update README.md**

Add the following sections/updates to `README.md`:

In the **セットアップ** section, add a detailed App ID guide before the existing config block:

```markdown
## アプリケーション ID の取得

1. [法人番号システム Web-API](https://www.houjin-bangou.nta.go.jp/webapi/) にアクセス
2. 「アプリケーションID の発行届出」から申請（無料）
3. メールアドレスを入力し、届出を送信
4. 届いたメールに記載されたアプリケーション ID を確認

発行されたアプリケーション ID は以下のいずれかの方法で設定できます：

### 環境変数（推奨）

```bash
export HOUJIN_APP_ID=your-application-id
```

### 設定ファイル

```bash
mkdir -p ~/.config/houjin
cat > ~/.config/houjin/config.yaml << EOF
app_id: your-application-id
format: json
EOF
```

優先順位: コマンドフラグ > 環境変数 (`HOUJIN_APP_ID`) > 設定ファイル (`~/.config/houjin/config.yaml`)
```

In **法人名で検索**, add new examples:

```markdown
# 法人種別でフィルタ（03=設立登記法人）
houjin search トヨタ --type 03

# 都道府県 + 市区町村コード指定
houjin search 株式会社 --pref 13 --city 101

# 全ページ自動取得
houjin search トヨタ --all

# 特定ページを指定
houjin search トヨタ --page 2
```

In **更新情報の取得**, add new examples:

```markdown
# 変更事由でフィルタ（01=新規）
houjin diff --from 2024-01-01 --to 2024-01-31 --kind 01

# 全ページ自動取得
houjin diff --from 2024-01-01 --to 2024-01-31 --all
```

In **法人番号で取得**, add batch examples:

```markdown
# ファイルから一括取得
houjin get --file numbers.txt

# 標準入力から取得
cat numbers.txt | houjin get --file -
echo "1180301018771" | houjin get --file -
```

Update the **コマンド一覧** table to include new flags:

| コマンド | 説明 |
|---------|------|
| `houjin get <番号...>` | 法人番号を指定して法人情報を取得 |
| `houjin get --file <ファイル>` | ファイルから法人番号を一括取得 |
| `houjin search <法人名>` | 法人名で検索 |
| `houjin diff` | 指定期間内の更新法人一覧を取得 |
| `houjin version` | バージョン情報を表示 |
| `houjin completion` | シェル補完スクリプトを生成 |

- [ ] **Step 2: Run build to ensure no compile errors**

Run: `cd /root/dev/planitai/houjin-cli && go build ./...`
Expected: Builds successfully.

- [ ] **Step 3: Commit**

```bash
git add README.md
git commit -m "docs: add App ID guide and document new features in README"
```

---

### Task 7: Final verification and PR

- [ ] **Step 1: Run full test suite**

Run: `cd /root/dev/planitai/houjin-cli && go test ./... -v -count=1`
Expected: All tests pass.

- [ ] **Step 2: Run linter**

Run: `cd /root/dev/planitai/houjin-cli && golangci-lint run ./...`
Expected: No issues.

- [ ] **Step 3: Verify build**

Run: `cd /root/dev/planitai/houjin-cli && go build -o houjin . && ./houjin --help`
Expected: Help output shows new flags.

- [ ] **Step 4: Verify subcommand help**

Run:
```bash
./houjin search --help
./houjin diff --help
./houjin get --help
```
Expected: Each shows new flags (`--page`, `--all`, `--type`, `--kind`, `--file`).

- [ ] **Step 5: Create PR**

```bash
git push origin <branch-name>
gh pr create --title "feat: paging, filters, batch input" --body "..."
```
