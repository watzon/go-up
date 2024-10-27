package widgets

import (
	"fmt"

	"github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/watzon/go-up/internal/types"
)

type StatsTable struct {
	*widgets.Table
}

func NewStatsTable() *StatsTable {
	table := widgets.NewTable()
	table.Border = false
	table.TextStyle = termui.NewStyle(termui.ColorWhite)
	table.RowSeparator = true
	table.FillRow = true
	table.ColumnWidths = []int{15, 15, 15, 15, 15}
	table.TextAlignment = termui.AlignCenter

	// Initialize with headers
	table.Rows = [][]string{
		{"Response", "Avg. Response", "Uptime", "Uptime", "Cert Exp."},
		{"(current)", "(24-hour)", "(24-hour)", "(30-day)", ""},
		{"--", "--", "--", "--", "--"},
	}

	return &StatsTable{Table: table}
}

func (s *StatsTable) Update(status types.ServiceStatus) {
	s.Rows = [][]string{
		{"Response", "Avg. Response", "Uptime", "Uptime", "Cert Exp."},
		{"(current)", "(24-hour)", "(24-hour)", "(30-day)", ""},
		{
			fmt.Sprintf("%d ms", status.ResponseTime),
			fmt.Sprintf("%d", int(status.AvgResponseTime)),
			fmt.Sprintf("%d%%", int(status.Uptime24Hours)),
			fmt.Sprintf("%d%%", int(status.Uptime30Days)),
			status.CertificateExpiry.Format("2006-01-02"),
		},
	}
}
