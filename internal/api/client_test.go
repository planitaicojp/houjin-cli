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
		if r.URL.Query().Get("id") == "" {
			http.Error(w, "missing id", http.StatusUnauthorized)
			return
		}
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

func TestClient_apiError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "forbidden", http.StatusForbidden)
	}))
	defer ts.Close()

	client := api.NewClient("test-app-id", api.WithBaseURL(ts.URL))
	_, err := client.GetByNumber([]string{"2180301018771"}, api.GetOptions{})
	if err == nil {
		t.Error("expected error for 403 response")
	}
}
