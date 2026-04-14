package output

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/olekukonko/tablewriter"
	"github.com/yehorkochetov/rey/internal/scanner"
)

var headerStyle = lipgloss.NewStyle().Bold(true)

func RenderGraveyard(resources []scanner.DeadResource) {
	table := tablewriter.NewWriter(os.Stdout)
	table.Header(
		headerStyle.Render("Type"),
		headerStyle.Render("Name/ID"),
		headerStyle.Render("Region"),
		headerStyle.Render("Age"),
		headerStyle.Render("Monthly Cost"),
	)

	for _, r := range resources {
		table.Append([]string{
			r.Type,
			nameOrID(r),
			r.Region,
			formatAge(r.Age),
			fmt.Sprintf("$%.2f/mo", r.MonthlyCost),
		})
	}

	table.Render()
}

func nameOrID(r scanner.DeadResource) string {
	if r.Name != "" {
		return r.Name
	}
	return r.ID
}

func formatAge(age time.Duration) string {
	if age <= 0 {
		return "-"
	}
	days := int(age.Hours() / 24)
	if days > 0 {
		return fmt.Sprintf("%dd", days)
	}
	return fmt.Sprintf("%dh", int(age.Hours()))
}
