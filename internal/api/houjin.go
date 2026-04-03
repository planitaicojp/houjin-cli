package api

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/planitaicojp/houjin-cli/internal/model"
)

// GetOptions configures the GetByNumber request.
type GetOptions struct {
	History bool
	Close   bool
}

// SearchOptions configures the SearchByName request.
type SearchOptions struct {
	Mode   string // "prefix" or "partial"
	Pref   string // prefecture code
	City   string // city code
	Close  bool
	Kind   string // corporation type filter (will be used by Task 2)
	Divide int    // page number (0 = default/first page)
}

// DiffOptions configures the GetDiff request.
type DiffOptions struct {
	Kind   string // change reason filter (01-99), used by Task 3
	Divide int    // page number (0 = default/first page)
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
	if opts.Kind != "" {
		params.Set("kind", opts.Kind)
	}
	params.Set("target", "1")

	if opts.Divide > 0 {
		params.Set("divide", strconv.Itoa(opts.Divide))
	}

	return c.fetch("name", params)
}

// GetDiff fetches corporations updated within the specified date range.
func (c *Client) GetDiff(from, to string, opts DiffOptions) (*model.Response, error) {
	params := url.Values{}
	params.Set("from", from)
	params.Set("to", to)
	if opts.Kind != "" {
		params.Set("kind", opts.Kind)
	}
	if opts.Divide > 0 {
		params.Set("divide", strconv.Itoa(opts.Divide))
	}
	return c.fetch("diff", params)
}

// SearchAllPages fetches all pages of a name search result.
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

// DiffAllPages fetches all pages of a diff result.
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
