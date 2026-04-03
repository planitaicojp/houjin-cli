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
