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
	Mode  string // "prefix" or "partial"
	Pref  string // prefecture code
	City  string // city code
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
