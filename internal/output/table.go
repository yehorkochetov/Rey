package output

import (
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/olekukonko/tablewriter"
	"github.com/yehorkochetov/rey/internal/scanner"
)

var (
	highCost = lipgloss.NewStyle().Foreground(lipgloss.Color("9")) // bright red
	medCost     = lipgloss.NewStyle().Foreground(lipgloss.Color("11")) // bright yellow
	defaultCost = lipgloss.NewStyle()
	footerStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10")) // bright green
	emptyStyle  = lipgloss.NewStyle().Faint(true).Italic(true)
)

func RenderGraveyard(resources []scanner.DeadResource) {
	if len(resources) == 0 {
		fmt.Println(emptyStyle.Render("No wasted resources found in this region"))
		return
	}

	sorted := make([]scanner.DeadResource, len(resources))
	copy(sorted, resources)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].Type < sorted[j].Type
	})

	table := tablewriter.NewWriter(os.Stdout)
	table.Header([]string{"Type", "Name/ID", "Region", "Age", "Monthly Cost"})

	var total float64
	for _, r := range sorted {
		style := costStyle(r.MonthlyCost)
		table.Append([]string{
			style.Render(r.Type),
			style.Render(nameOrID(r)),
			style.Render(r.Region),
			style.Render(formatAge(r.Age)),
			style.Render(fmt.Sprintf("$%.2f/mo", r.MonthlyCost)),
		})
		total += r.MonthlyCost
	}

	table.Render()

	fmt.Println(footerStyle.Render(
		fmt.Sprintf("Found %d resources wasting $%.2f/month", len(resources), total),
	))
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
