package main

import (
        "fmt"
        "strings"

        "github.com/charmbracelet/bubbles/table"
        "github.com/charmbracelet/bubbles/textinput"
        "github.com/charmbracelet/bubbletea"
        "github.com/charmbracelet/lipgloss"
)

// ViewMode represents the current view state
type ViewMode int

const (
        ViewTable ViewMode = iota
        ViewEdit
        ViewHistory
        ViewHelp
)

// Model represents the application state
type Model struct {
        mode        ViewMode
        table       table.Model
        jobs        []CronJob
        selected    int
        editing     bool
        editingJob  CronJob
        editIndex   int
        inputs      []textinput.Model
        activeInput int
        history     []LogEntry
        error       string
        message     string
}

// Styles
var (
        baseStyle = lipgloss.NewStyle().
                BorderStyle(lipgloss.NormalBorder()).
                BorderForeground(lipgloss.Color("240"))

        titleStyle = lipgloss.NewStyle().
                Foreground(lipgloss.Color("86")).
                Bold(true).
                Margin(0, 0, 1, 0)

        errorStyle = lipgloss.NewStyle().
                Foreground(lipgloss.Color("196")).
                Bold(true)

        successStyle = lipgloss.NewStyle().
                Foreground(lipgloss.Color("46")).
                Bold(true)

        helpStyle = lipgloss.NewStyle().
                Foreground(lipgloss.Color("244"))

        keybindingStyle = lipgloss.NewStyle().
                Foreground(lipgloss.Color("241")).
                MarginTop(1)

        cronDescStyle = lipgloss.NewStyle().
                Foreground(lipgloss.Color("208")).
                Italic(true)
)

// NewModel creates a new application model
func NewModel() Model {
        // Create table
        columns := []table.Column{
                {Title: "Description", Width: 25},
                {Title: "Cron Expression", Width: 15},
                {Title: "Next Run", Width: 20},
                {Title: "Last Run", Width: 20},
                {Title: "Command", Width: 30},
        }

        t := table.New(
                table.WithColumns(columns),
                table.WithFocused(true),
                table.WithHeight(15),
        )

        s := table.DefaultStyles()
        s.Header = s.Header.
                BorderStyle(lipgloss.NormalBorder()).
                BorderForeground(lipgloss.Color("240")).
                BorderBottom(true).
                Bold(false)
        s.Selected = s.Selected.
                Foreground(lipgloss.Color("229")).
                Background(lipgloss.Color("57")).
                Bold(false)
        t.SetStyles(s)

        // Create text inputs for editing
        inputs := make([]textinput.Model, 4)
        
        // Description input
        inputs[0] = textinput.New()
        inputs[0].Placeholder = "Job description..."
        inputs[0].Focus()
        inputs[0].CharLimit = 100
        inputs[0].Width = 50

        // Cron expression input
        inputs[1] = textinput.New()
        inputs[1].Placeholder = "0 9 * * *"
        inputs[1].CharLimit = 50
        inputs[1].Width = 20

        // Command input
        inputs[2] = textinput.New()
        inputs[2].Placeholder = "/path/to/script.sh"
        inputs[2].CharLimit = 200
        inputs[2].Width = 60

        // Log file input
        inputs[3] = textinput.New()
        inputs[3].Placeholder = "job_name"
        inputs[3].CharLimit = 50
        inputs[3].Width = 30

        m := Model{
                mode:        ViewTable,
                table:       t,
                inputs:      inputs,
                activeInput: 0,
        }

        // Load cron jobs
        m.loadJobs()

        return m
}

// loadJobs loads cron jobs from the system
func (m *Model) loadJobs() {
        jobs, err := ReadCrontab()
        if err != nil {
                m.error = fmt.Sprintf("Error loading cron jobs: %v", err)
                return
        }

        // Update last run times from log files
        for i := range jobs {
                if jobs[i].LogFile != "" {
                        jobs[i].LastRun = GetLastRunFromLogFile(jobs[i].LogFile)
                }
        }

        m.jobs = jobs
        m.updateTable()
        m.error = ""
}

// updateTable refreshes the table with current job data
func (m *Model) updateTable() {
        rows := make([]table.Row, len(m.jobs))
        for i, job := range m.jobs {
                description := job.Description
                if description == "" {
                        description = "No description"
                }

                nextRun := "Never"
                if !job.NextRun.IsZero() {
                        nextRun = job.NextRun.Format("Jan 2, 15:04")
                }

                var lastRun string
                if job.LogFile == "" {
                        lastRun = "-"
                } else if !job.LastRun.IsZero() {
                        lastRun = job.LastRun.Format("Jan 2, 15:04")
                } else {
                        lastRun = "Never"
                }

                // Strip logging from command for display
                command := StripLoggingFromCommand(job.Command)
                if len(command) > 28 {
                        command = command[:25] + "..."
                }

                rows[i] = table.Row{
                        description,
                        job.Expression,
                        nextRun,
                        lastRun,
                        command,
                }
        }

        m.table.SetRows(rows)
}

// Init implements the tea.Model interface
func (m Model) Init() tea.Cmd {
        return textinput.Blink
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
        var cmd tea.Cmd

        switch msg := msg.(type) {
        case tea.KeyMsg:
                switch m.mode {
                case ViewTable:
                        return m.updateTableView(msg)
                case ViewEdit:
                        return m.updateEdit(msg)
                case ViewHistory:
                        return m.updateHistory(msg)
                case ViewHelp:
                        return m.updateHelp(msg)
                }

        case tea.WindowSizeMsg:
                m.table.SetWidth(msg.Width - 4)
                m.table.SetHeight(msg.Height - 10)
        }

        return m, cmd
}

// updateTableView handles key presses in table view
func (m Model) updateTableView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
        var cmd tea.Cmd

        switch msg.String() {
        case "q", "ctrl+c":
                return m, tea.Quit

        case "n":
                m.mode = ViewEdit
                m.editing = false
                m.editIndex = -1
                m.editingJob = CronJob{}
                m.resetInputs()
                return m, textinput.Blink

        case "e":
                if len(m.jobs) > 0 {
                        m.mode = ViewEdit
                        m.editing = true
                        m.selected = m.table.Cursor()
                        m.editIndex = m.selected
                        if m.editIndex < len(m.jobs) {
                                m.editingJob = m.jobs[m.editIndex]
                                m.populateInputs()
                        }
                }
                return m, textinput.Blink

        case "h":
                if len(m.jobs) > 0 {
                        m.selected = m.table.Cursor()
                        if m.selected < len(m.jobs) {
                                m.mode = ViewHistory
                                m.history = GetJobHistoryFromLogFile(m.jobs[m.selected].LogFile)
                        }
                }
                return m, nil

        case "r":
                m.loadJobs()
                m.message = "Refreshed cron jobs"
                return m, nil
        }

        m.table, cmd = m.table.Update(msg)
        return m, cmd
}

// updateEdit handles key presses in edit view
func (m Model) updateEdit(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
        var cmd tea.Cmd

        switch msg.String() {
        case "ctrl+c":
                m.mode = ViewTable
                m.error = ""
                return m, nil

        case "ctrl+s":
                return m.saveJob()

        case "tab", "shift+tab", "enter", "up", "down":
                s := msg.String()

                if s == "up" || s == "shift+tab" {
                        m.activeInput--
                } else {
                        m.activeInput++
                }

                if m.activeInput > len(m.inputs)-1 {
                        m.activeInput = 0
                } else if m.activeInput < 0 {
                        m.activeInput = len(m.inputs) - 1
                }

                for i := 0; i < len(m.inputs); i++ {
                        if i == m.activeInput {
                                m.inputs[i].Focus()
                        } else {
                                m.inputs[i].Blur()
                        }
                }

                return m, textinput.Blink

        case "ctrl+/":
                m.mode = ViewHelp
                return m, nil
        }

        m.inputs[m.activeInput], cmd = m.inputs[m.activeInput].Update(msg)
        return m, cmd
}

// updateHistory handles key presses in history view
func (m Model) updateHistory(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
        switch msg.String() {
        case "esc", "q":
                m.mode = ViewTable
                return m, nil
        }
        return m, nil
}

// updateHelp handles key presses in help view
func (m Model) updateHelp(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
        switch msg.String() {
        case "esc", "q":
                m.mode = ViewEdit
                return m, nil
        }
        return m, nil
}

// saveJob saves the current job being edited
func (m Model) saveJob() (tea.Model, tea.Cmd) {
        // Validate inputs
        description := m.inputs[0].Value()
        expression := m.inputs[1].Value()
        command := m.inputs[2].Value()
        logFile := m.inputs[3].Value()

        if expression == "" {
                m.error = "Cron expression is required"
                return m, nil
        }

        if command == "" {
                m.error = "Command is required"
                return m, nil
        }

        // Log file is optional - leave empty for no logging

        if err := ValidateCronExpression(expression); err != nil {
                m.error = fmt.Sprintf("Invalid cron expression: %v", err)
                return m, nil
        }

        // Create the job
        nextRun, _ := GetNextRunTime(expression)
        job := CronJob{
                Description: description,
                Expression:  expression,
                Command:     command,
                LogFile:     logFile,
                NextRun:     nextRun,
        }

        // Add or update job
        if m.editing && m.editIndex >= 0 && m.editIndex < len(m.jobs) {
                m.jobs[m.editIndex] = job
        } else {
                m.jobs = append(m.jobs, job)
        }

        // Save to crontab
        if err := WriteCrontab(m.jobs); err != nil {
                m.error = fmt.Sprintf("Error saving crontab: %v", err)
                return m, nil
        }

        m.mode = ViewTable
        m.updateTable()
        m.message = "Job saved successfully"
        m.error = ""

        return m, nil
}

// resetInputs clears all input fields
func (m *Model) resetInputs() {
        for i := range m.inputs {
                m.inputs[i].SetValue("")
        }
        m.activeInput = 0
        m.inputs[0].Focus()
        for i := 1; i < len(m.inputs); i++ {
                m.inputs[i].Blur()
        }
}

// populateInputs fills input fields with current job data
func (m *Model) populateInputs() {
        m.inputs[0].SetValue(m.editingJob.Description)
        m.inputs[1].SetValue(m.editingJob.Expression)
        m.inputs[2].SetValue(m.editingJob.Command)
        m.inputs[3].SetValue(m.editingJob.LogFile)
        
        m.activeInput = 0
        m.inputs[0].Focus()
        for i := 1; i < len(m.inputs); i++ {
                m.inputs[i].Blur()
        }
}

// View renders the current view
func (m Model) View() string {
        switch m.mode {
        case ViewTable:
                return m.viewTable()
        case ViewEdit:
                return m.viewEdit()
        case ViewHistory:
                return m.viewHistory()
        case ViewHelp:
                return m.viewHelp()
        default:
                return "Unknown view"
        }
}

// viewTable renders the main table view
func (m Model) viewTable() string {
        var b strings.Builder

        // Title
        b.WriteString(titleStyle.Render("TUI Cron Job Manager"))
        b.WriteString("\n\n")

        // Error or success message
        if m.error != "" {
                b.WriteString(errorStyle.Render("Error: " + m.error))
                b.WriteString("\n\n")
        } else if m.message != "" {
                b.WriteString(successStyle.Render(m.message))
                b.WriteString("\n\n")
        }

        // Center the table
        tableView := m.table.View()
        centeredTable := lipgloss.NewStyle().
                Width(120).
                Align(lipgloss.Center).
                Render(tableView)
        b.WriteString(baseStyle.Render(centeredTable))
        b.WriteString("\n")

        // Keybindings
        keybindings := []string{
                "n: new job",
                "e: edit job", 
                "h: job history",
                "r: refresh",
                "q: quit",
        }
        b.WriteString(keybindingStyle.Render(strings.Join(keybindings, " • ")))

        return b.String()
}

// viewEdit renders the job editing view
func (m Model) viewEdit() string {
        var b strings.Builder

        // Title with purple background like in image
        titleBox := lipgloss.NewStyle().
                Background(lipgloss.Color("99")).
                Foreground(lipgloss.Color("15")).
                Padding(0, 1).
                Bold(true).
                Render("Edit Job")
        b.WriteString(titleBox)
        b.WriteString("\n\n")

        // Error message
        if m.error != "" {
                b.WriteString(errorStyle.Render("Error: " + m.error))
                b.WriteString("\n\n")
        }

        // Description field
        b.WriteString("Description:")
        b.WriteString("\n")
        
        // Style the description input with border
        descBorderStyle := lipgloss.NewStyle().
                Border(lipgloss.NormalBorder()).
                BorderForeground(lipgloss.Color("240"))
        if m.activeInput == 0 {
                descBorderStyle = descBorderStyle.BorderForeground(lipgloss.Color("86"))
        }
        descInput := descBorderStyle.Width(60).Padding(0, 1).Render(m.inputs[0].View())
        b.WriteString(descInput)
        b.WriteString("\n\n")

        // Cron expression field
        b.WriteString("Cron Expression:")
        b.WriteString("\n")
        
        // Style the cron input with border (smaller)
        cronBorderStyle := lipgloss.NewStyle().
                Border(lipgloss.NormalBorder()).
                BorderForeground(lipgloss.Color("240"))
        if m.activeInput == 1 {
                cronBorderStyle = cronBorderStyle.BorderForeground(lipgloss.Color("86"))
        }
        cronInput := cronBorderStyle.Width(20).Padding(0, 1).Render(m.inputs[1].View())
        
        // Show human-readable description if expression is valid
        cronDesc := ""
        if m.inputs[1].Value() != "" {
                if err := ValidateCronExpression(m.inputs[1].Value()); err == nil {
                        description := ParseCronExpression(m.inputs[1].Value())
                        cronDesc = " → " + description
                }
        }
        
        cronLine := cronInput + cronDescStyle.Render(cronDesc)
        b.WriteString(cronLine)
        b.WriteString("\n\n")

        // Command field
        b.WriteString("Command:")
        b.WriteString("\n")
        
        // Style the command input with border
        cmdBorderStyle := lipgloss.NewStyle().
                Border(lipgloss.NormalBorder()).
                BorderForeground(lipgloss.Color("240"))
        if m.activeInput == 2 {
                cmdBorderStyle = cmdBorderStyle.BorderForeground(lipgloss.Color("86"))
        }
        cmdInput := cmdBorderStyle.Width(80).Padding(0, 1).Render(m.inputs[2].View())
        b.WriteString(cmdInput)
        b.WriteString("\n\n")

        // Log file field
        b.WriteString("Log File:")
        b.WriteString("\n")
        
        // Style the log file input with border
        logBorderStyle := lipgloss.NewStyle().
                Border(lipgloss.NormalBorder()).
                BorderForeground(lipgloss.Color("240"))
        if m.activeInput == 3 {
                logBorderStyle = logBorderStyle.BorderForeground(lipgloss.Color("86"))
        }
        logInput := logBorderStyle.Width(30).Padding(0, 1).Render(m.inputs[3].View())
        logDesc := cronDescStyle.Render(" (saved as ~/.cron_history/[name].log)")
        b.WriteString(logInput + logDesc)
        b.WriteString("\n\n")

        // Keybindings
        keybindings := []string{
                "ctrl+s: save",
                "ctrl+c: cancel", 
                "tab: next field",
                "ctrl+/: cron help",
        }
        b.WriteString(keybindingStyle.Render(strings.Join(keybindings, " • ")))

        return b.String()
}

// viewHistory renders the job history view
func (m Model) viewHistory() string {
        var b strings.Builder

        // Title
        job := m.jobs[m.selected]
        b.WriteString(titleStyle.Render(fmt.Sprintf("Log History: %s", job.Description)))
        b.WriteString("\n")
        b.WriteString(helpStyle.Render(fmt.Sprintf("Command: %s", StripLoggingFromCommand(job.Command))))
        b.WriteString("\n")
        if job.LogFile != "" {
                b.WriteString(helpStyle.Render(fmt.Sprintf("Log File: ~/.cron_history/%s.log", job.LogFile)))
                b.WriteString("\n")
        }
        b.WriteString("\n")

        if job.LogFile == "" {
                b.WriteString(helpStyle.Render("No log file configured for this job."))
                b.WriteString("\n")
                b.WriteString(helpStyle.Render("Edit the job and add a log file name to enable logging."))
        } else if len(m.history) == 0 {
                b.WriteString(helpStyle.Render("No log entries found"))
        } else {
                // Log entries with color coding
                for i, entry := range m.history {
                        if i >= 50 { // Show more entries since it's from log files
                                break
                        }

                        // Color code based on content
                        line := entry.Message
                        if strings.Contains(strings.ToLower(line), "error") {
                                line = errorStyle.Render(line)
                        } else if strings.Contains(strings.ToLower(line), "warning") {
                                line = cronDescStyle.Render(line)
                        } else if strings.Contains(line, "Starting job") {
                                line = successStyle.Render(line)
                        }

                        b.WriteString(line)
                        b.WriteString("\n")
                }
        }

        b.WriteString("\n")

        // Keybindings
        keybindings := []string{
                "Esc/q: back to jobs",
        }
        b.WriteString(keybindingStyle.Render(strings.Join(keybindings, " • ")))

        return baseStyle.Render(b.String())
}
