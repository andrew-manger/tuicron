# TUI Cron Job Manager

## Project Overview
A terminal-based cron job manager built with Go and the Bubbletea TUI framework. This application provides an intuitive interface for managing cron jobs with table navigation, editing capabilities, and history viewing.

## Architecture

### Core Components
- **main.go**: Entry point and application initialization
- **ui.go**: Main UI logic using Bubbletea framework with multiple view modes
- **cron.go**: Cron job parsing, validation, and system interaction
- **logs.go**: System log parsing for job execution history
- **help.go**: Help system with cron expression documentation

### Dependencies
- `github.com/charmbracelet/bubbletea`: TUI framework
- `github.com/charmbracelet/lipgloss`: Styling and layout
- `github.com/charmbracelet/bubbles`: UI components (table, textinput)
- `github.com/robfig/cron/v3`: Cron expression parsing and validation
- `github.com/olekukonko/tablewriter`: Table formatting

## Features Implemented

### Main Interface
- **Centered Table View**: Displays cron jobs in a centered, structured table with columns:
  - Description (user-provided)
  - Cron Expression 
  - Next Run Time (calculated)
  - Last Run Time (from system logs)
  - Command

### Navigation & Controls
- **Arrow Keys**: Navigate through the job list
- **Keyboard Shortcuts**:
  - `n`: Add new cron job
  - `e`: Edit selected job
  - `h`: View execution history for selected job
  - `d`: Delete selected job (with confirmation)
  - `r`: Refresh job list
  - `q`: Quit application

### Edit Mode
- **Visual Design**: Purple "Edit Job" header with bordered input fields matching terminal aesthetics
- **Field Navigation**: Tab/Shift+Tab to move between fields with highlighted active borders
- **Form Fields**:
  - Description input (wide bordered field)
  - Cron expression input (compact field) with real-time human-readable translation displayed inline
  - Command input (full-width bordered field)
  - Log file input (compact field) - creates ~/.cron_history/[name].log for job output
- **Help System**: Ctrl+/ opens cron expression help
- **Save/Cancel**: Ctrl+S to save, Ctrl+C to cancel

### Cron Expression Features
- **Validation**: Real-time validation of cron expressions
- **Human-Readable**: Converts cron expressions to plain English
- **Help Documentation**: Comprehensive guide with examples and special characters

### Delete Functionality
- **Safety-First Design**: Confirmation dialog prevents accidental deletion
- **Job Preview**: Shows complete job details before deletion including description, cron expression (with human-readable translation), and command
- **Default to Safe**: "No" option is selected by default to prevent accidental deletions
- **Intuitive Navigation**: Left/right arrow keys to choose between "No" and "Yes", Enter to confirm, Esc to cancel
- **Visual Feedback**: Clear color coding with green border for "No" and red border for "Yes" when selected

### Log Management & History Viewing
- **Dedicated Log Files**: Each job creates a log file in ~/.cron_history/[name].log
- **Automatic Output Capture**: Commands are modified to capture output with timestamps
- **Command Display**: Table shows clean commands without logging redirection
- **Log File History**: View complete log contents with color-coded entries
  - Green: Job start messages
  - Orange: Warning messages  
  - Red: Error messages
- **Last Run Detection**: Parses log files for most recent execution timestamps

## Technical Implementation

### Crontab Integration & Logging
- **Real Crontab Loading**: Loads actual crontab contents on startup, preserving jobs added outside the TUI
- **Command Processing**: Automatically adds logging redirection (`>> /path/to/logfile.log 2>&1`) to commands
- **Smart Parsing**: Extracts clean commands and log file names from existing cron entries
- **Log Directory Management**: Creates ~/.cron_history/ directory automatically
- **Smart Fallback**: When crontab is empty or cron daemon not running, displays sample data with log files:
  - Daily backup script (backup.log)
  - Weekly system update (system_update.log)
  - Hourly temp file cleanup (cleanup.log)
- **External Changes**: Refreshes crontab data and log file timestamps

### Error Handling
- Graceful fallback when crontab is unavailable
- Input validation for cron expressions
- Safe file operations with backup creation

### Styling
- Professional color scheme with syntax highlighting
- Responsive layout that adapts to terminal size
- Clear visual hierarchy and status indicators

## User Preferences
- Interface Language: English
- Error Handling: Graceful fallback with sample data
- Input Validation: Real-time with helpful error messages

## Recent Changes
- **August 2025**: Advanced logging system, UI improvements, and delete functionality
  - **Logging System**: Added dedicated log file field in edit mode
    - Creates ~/.cron_history/[name].log for each job
    - Automatically appends logging redirection to cron commands
    - Strips logging display from table view for cleaner interface
    - Parses log files for accurate last run timestamps
  - **Enhanced History View**: Shows complete log file contents with color coding
  - **UI Improvements**: 
    - Centered table display for better visual balance
    - Redesigned edit interface with bordered fields and log file input
    - Changed help shortcut from Ctrl+? to Ctrl+/ for better accessibility
    - Improved keybinding display (combined Esc/q shortcuts)
  - **Delete Functionality**: Added comprehensive job deletion with safety confirmation
    - 'd' keybinding to delete selected job with confirmation dialog
    - Safety-first design with "No" as default selection to prevent accidents
    - Job preview shows complete details before deletion
    - Arrow key navigation for Yes/No selection with color-coded feedback
    - Proper crontab integration to save changes immediately
  - **Crontab Integration**: Enhanced loading to preserve external changes
- **December 2024**: Initial implementation completed
- **Architecture**: Built modular codebase with separation of concerns
- **UI Framework**: Implemented full Bubbletea-based interface
- **Sample Mode**: Added fallback data for environments without cron
- **Validation**: Integrated robust cron expression validation
- **Help System**: Created comprehensive cron syntax documentation

## Development Notes
- Built for Go 1.19 compatibility
- Uses Nix package management for dependencies
- Console-based TUI application (not web-based)
- Follows Unix philosophy with simple, focused functionality

## Usage Instructions
Run `go run .` or `./tuicron` to start the application. The interface is self-explanatory with keyboard shortcuts displayed at the bottom of each view.
