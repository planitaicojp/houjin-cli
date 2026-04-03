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
