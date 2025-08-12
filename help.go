package main

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// viewHelp renders the help view with cron expression information
func (m Model) viewHelp() string {
	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render("Cron Expression Help"))
	b.WriteString("\n\n")

	// Format explanation
	b.WriteString(lipgloss.NewStyle().Bold(true).Render("Cron Expression Format:"))
	b.WriteString("\n")
	b.WriteString("* * * * *")
	b.WriteString("\n")
	b.WriteString("│ │ │ │ │")
	b.WriteString("\n")
	b.WriteString("│ │ │ │ └─── day of week (0-7, 0 or 7 = Sunday)")
	b.WriteString("\n")
	b.WriteString("│ │ │ └───── month (1-12)")
	b.WriteString("\n")
	b.WriteString("│ │ └─────── day of month (1-31)")
	b.WriteString("\n")
	b.WriteString("│ └───────── hour (0-23)")
	b.WriteString("\n")
	b.WriteString("└─────────── minute (0-59)")
	b.WriteString("\n\n")

	// Special characters
	b.WriteString(lipgloss.NewStyle().Bold(true).Render("Special Characters:"))
	b.WriteString("\n")
	b.WriteString("*     - Any value (wildcard)")
	b.WriteString("\n")
	b.WriteString(",     - Value list separator (e.g., 1,3,5)")
	b.WriteString("\n")
	b.WriteString("-     - Range of values (e.g., 1-5)")
	b.WriteString("\n")
	b.WriteString("/     - Step values (e.g., */5 = every 5)")
	b.WriteString("\n\n")

	// Common examples
	b.WriteString(lipgloss.NewStyle().Bold(true).Render("Common Examples:"))
	b.WriteString("\n")

	examples := [][]string{
		{"0 0 * * *", "Daily at midnight"},
		{"0 9 * * *", "Daily at 9:00 AM"},
		{"0 9 * * 1", "Every Monday at 9:00 AM"},
		{"0 9-17 * * *", "Every hour from 9 AM to 5 PM"},
		{"*/15 * * * *", "Every 15 minutes"},
		{"0 */2 * * *", "Every 2 hours"},
		{"0 9 1 * *", "First day of every month at 9:00 AM"},
		{"0 9 * * 1-5", "Weekdays (Mon-Fri) at 9:00 AM"},
		{"0 0 1 1 *", "New Year's Day at midnight"},
		{"0 12 * * 0", "Every Sunday at noon"},
	}

	for _, example := range examples {
		b.WriteString(cronDescStyle.Render(example[0]))
		b.WriteString("  ")
		b.WriteString(example[1])
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// Presets
	b.WriteString(lipgloss.NewStyle().Bold(true).Render("Special Presets (if supported):"))
	b.WriteString("\n")
	b.WriteString("@yearly   - Same as 0 0 1 1 *")
	b.WriteString("\n")
	b.WriteString("@monthly  - Same as 0 0 1 * *")
	b.WriteString("\n")
	b.WriteString("@weekly   - Same as 0 0 * * 0")
	b.WriteString("\n")
	b.WriteString("@daily    - Same as 0 0 * * *")
	b.WriteString("\n")
	b.WriteString("@hourly   - Same as 0 * * * *")
	b.WriteString("\n\n")

	// Tips
	b.WriteString(lipgloss.NewStyle().Bold(true).Render("Tips:"))
	b.WriteString("\n")
	b.WriteString("• Use absolute paths for commands and scripts")
	b.WriteString("\n")
	b.WriteString("• Test your cron expressions before saving")
	b.WriteString("\n")
	b.WriteString("• Consider timezone differences")
	b.WriteString("\n")
	b.WriteString("• Use >> /path/to/logfile 2>&1 for logging")
	b.WriteString("\n\n")

	// Keybindings
	keybindings := []string{
		"Esc: back to edit",
		"q: back to edit",
	}
	b.WriteString(keybindingStyle.Render(strings.Join(keybindings, " • ")))

	return baseStyle.Render(b.String())
}
