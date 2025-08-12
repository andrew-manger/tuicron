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
        inputs := make([]textinput.Model, 3)
        
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

                lastRun := "Never"
                if !job.LastRun.IsZero() {
                        lastRun = job.LastRun.Format("Jan 2, 15:04")
                }

                command := job.Command
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
                                m.history = GetJobHistory(m.jobs[m.selected].Command)
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

        case "ctrl+?":
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

        if expression == "" {
                m.error = "Cron expression is required"
                return m, nil
        }

        if command == "" {
                m.error = "Command is required"
                return m, nil
        }

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

        // Table
        b.WriteString(baseStyle.Render(m.table.View()))
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

        // Title
        title := "Add New Cron Job"
        if m.editing {
                title = "Edit Cron Job"
        }
        b.WriteString(titleStyle.Render(title))
        b.WriteString("\n\n")

        // Error message
        if m.error != "" {
                b.WriteString(errorStyle.Render("Error: " + m.error))
                b.WriteString("\n\n")
        }

        // Description field
        b.WriteString("Description:")
        b.WriteString("\n")
        b.WriteString(m.inputs[0].View())
        b.WriteString("\n\n")

        // Cron expression field
        b.WriteString("Cron Expression:")
        b.WriteString("\n")
        b.WriteString(m.inputs[1].View())
        
        // Show human-readable description if expression is valid
        if m.inputs[1].Value() != "" {
                if err := ValidateCronExpression(m.inputs[1].Value()); err == nil {
                        description := ParseCronExpression(m.inputs[1].Value())
                        b.WriteString("  ")
                        b.WriteString(cronDescStyle.Render("(" + description + ")"))
                }
        }
        b.WriteString("\n\n")

        // Command field
        b.WriteString("Command:")
        b.WriteString("\n")
        b.WriteString(m.inputs[2].View())
        b.WriteString("\n\n")

        // Keybindings
        keybindings := []string{
                "Tab: navigate fields",
                "Ctrl+S: save",
                "Ctrl+C: cancel",
                "Ctrl+?: help",
        }
        b.WriteString(keybindingStyle.Render(strings.Join(keybindings, " • ")))

        return baseStyle.Render(b.String())
}

// viewHistory renders the job history view
func (m Model) viewHistory() string {
        var b strings.Builder

        // Title
        job := m.jobs[m.selected]
        b.WriteString(titleStyle.Render(fmt.Sprintf("Execution History: %s", job.Description)))
        b.WriteString("\n")
        b.WriteString(helpStyle.Render(fmt.Sprintf("Command: %s", job.Command)))
        b.WriteString("\n\n")

        if len(m.history) == 0 {
                b.WriteString(helpStyle.Render("No execution history found"))
        } else {
                // History entries
                for i, entry := range m.history {
                        if i >= 20 { // Limit display to 20 most recent entries
                                break
                        }

                        timestamp := entry.Timestamp.Format("2006-01-02 15:04:05")
                        status := entry.Status
                        message := entry.Message

                        b.WriteString(fmt.Sprintf("%s [%s] %s", timestamp, status, message))
                        b.WriteString("\n")
                }
        }

        b.WriteString("\n")

        // Keybindings
        keybindings := []string{
                "Esc: back to jobs",
                "q: back to jobs",
        }
        b.WriteString(keybindingStyle.Render(strings.Join(keybindings, " • ")))

        return baseStyle.Render(b.String())
}
