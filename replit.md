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
  - `r`: Refresh job list
  - `q`: Quit application

### Edit Mode
- **Visual Design**: Purple "Edit Job" header with bordered input fields matching terminal aesthetics
- **Field Navigation**: Tab/Shift+Tab to move between fields with highlighted active borders
- **Form Fields**:
  - Description input (wide bordered field)
  - Cron expression input (compact field) with real-time human-readable translation displayed inline
  - Command input (full-width bordered field)
- **Help System**: Ctrl+/ opens cron expression help
- **Save/Cancel**: Ctrl+S to save, Ctrl+C to cancel

### Cron Expression Features
- **Validation**: Real-time validation of cron expressions
- **Human-Readable**: Converts cron expressions to plain English
- **Help Documentation**: Comprehensive guide with examples and special characters

### History Viewing
- **Execution Logs**: Shows job execution history from system logs
- **Multiple Sources**: Checks systemd journal, syslog, and cron logs
- **Formatted Display**: Clean presentation of timestamps and status

## Technical Implementation

### Crontab Integration & Sample Data
- **Real Crontab Loading**: Application now loads actual crontab contents on startup, preserving jobs added outside the TUI
- **Smart Fallback**: When crontab is empty or cron daemon is not running (common in containerized environments), displays sample data for demonstration:
  - Daily backup script
  - Weekly system update  
  - Hourly temp file cleanup
- **External Changes**: Refreshes crontab data to show jobs created by other tools

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
- **August 2025**: UI improvements and crontab integration
  - Centered table display for better visual balance
  - Redesigned edit interface with bordered fields matching terminal aesthetics
  - Changed help shortcut from Ctrl+? to Ctrl+/ for better accessibility
  - Improved keybinding display (combined Esc/q shortcuts)
  - Enhanced crontab loading to preserve external changes
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