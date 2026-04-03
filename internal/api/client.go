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
func WithBaseURL(u string) Option {
	return func(c *Client) { c.baseURL = u }
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
