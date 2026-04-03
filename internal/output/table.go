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
