package output

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/olekukonko/tablewriter"
	"github.com/yehorkochetov/rey/internal/scanner"
)

var (
	headerStyle = lipgloss.NewStyle().Bold(true)
	highCost    = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))  // bright red
	medCost     = lipgloss.NewStyle().Foreground(lipgloss.Color("11")) // bright yellow
	defaultCost = lipgloss.NewStyle()
)

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
		style := costStyle(r.MonthlyCost)
		table.Append([]string{
			style.Render(r.Type),
			style.Render(nameOrID(r)),
			style.Render(r.Region),
			style.Render(formatAge(r.Age)),
			style.Render(fmt.Sprintf("$%.2f/mo", r.MonthlyCost)),
		})
	}

	table.Render()
}

func costStyle(cost float64) lipgloss.Style {
	switch {
	case cost > 10:
		return highCost
	case cost > 1:
		return medCost
	default:
		return defaultCost
	}
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
