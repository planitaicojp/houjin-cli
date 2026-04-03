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
