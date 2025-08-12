package main

import (
        "bufio"
        "fmt"
        "os"
        "sort"
        "strings"
        "time"
)

// getHistoryDir returns the path to the cron history directory
func getHistoryDir() string {
        homeDir, err := os.UserHomeDir()
        if err != nil {
                return "/tmp/.cron_history" // fallback
        }
        return fmt.Sprintf("%s/.cron_history", homeDir)
}

// ensureHistoryDir creates the history directory if it doesn't exist
func ensureHistoryDir() error {
        historyDir := getHistoryDir()
        return os.MkdirAll(historyDir, 0755)
}

// getCommandHash creates a safe filename from a command
func getCommandHash(command string) string {
        // Replace unsafe characters with underscores
        safe := strings.ReplaceAll(command, "/", "_")
        safe = strings.ReplaceAll(safe, " ", "_")
        safe = strings.ReplaceAll(safe, "\\", "_")
        safe = strings.ReplaceAll(safe, ":", "_")
        safe = strings.ReplaceAll(safe, "*", "_")
        safe = strings.ReplaceAll(safe, "?", "_")
        safe = strings.ReplaceAll(safe, "\"", "_")
        safe = strings.ReplaceAll(safe, "<", "_")
        safe = strings.ReplaceAll(safe, ">", "_")
        safe = strings.ReplaceAll(safe, "|", "_")
        
        // Limit length and add suffix
        if len(safe) > 50 {
                safe = safe[:50]
        }
        return safe + ".log"
}

// LogJobExecution records a job execution to our custom history
func LogJobExecution(command, status, message string) error {
        if err := ensureHistoryDir(); err != nil {
                return err
        }
        
        historyDir := getHistoryDir()
        filename := getCommandHash(command)
        filepath := fmt.Sprintf("%s/%s", historyDir, filename)
        
        file, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
        if err != nil {
                return err
        }
        defer file.Close()
        
        timestamp := time.Now().Format("2006-01-02 15:04:05")
        logLine := fmt.Sprintf("%s|%s|%s|%s\n", timestamp, status, command, message)
        
        _, err = file.WriteString(logLine)
        return err
}

// getHistoryFromFile reads job history from our custom log file
func getHistoryFromFile(command string) []LogEntry {
        var entries []LogEntry
        
        historyDir := getHistoryDir()
        filename := getCommandHash(command)
        filepath := fmt.Sprintf("%s/%s", historyDir, filename)
        
        file, err := os.Open(filepath)
        if err != nil {
                return entries // file doesn't exist yet
        }
        defer file.Close()
        
        scanner := bufio.NewScanner(file)
        for scanner.Scan() {
                line := scanner.Text()
                parts := strings.SplitN(line, "|", 4)
                if len(parts) != 4 {
                        continue
                }
                
                timestamp, err := time.Parse("2006-01-02 15:04:05", parts[0])
                if err != nil {
                        continue
                }
                
                entry := LogEntry{
                        Timestamp: timestamp,
                        Status:    parts[1],
                        Message:   parts[3],
                }
                entries = append(entries, entry)
        }
        
        // Return newest first
        sort.Slice(entries, func(i, j int) bool {
                return entries[i].Timestamp.After(entries[j].Timestamp)
        })
        
        return entries
}

// getLastRunFromHistory gets the most recent execution time from our history
func getLastRunFromHistory(command string) time.Time {
        entries := getHistoryFromFile(command)
        if len(entries) > 0 {
                return entries[0].Timestamp
        }
        return time.Time{}
}

// CreateSampleHistory creates some sample history entries for demonstration
func CreateSampleHistory() {
        // Create sample history for our demo jobs
        commands := []string{
                "/home/user/scripts/backup.sh",
                "sudo apt update && sudo apt upgrade -y", 
                "find /tmp -type f -mtime +1 -delete",
        }
        
        for _, cmd := range commands {
                // Create entries for the last few days
                for i := 1; i <= 5; i++ {
                        pastTime := time.Now().Add(-time.Duration(i*24) * time.Hour)
                        LogJobExecution(cmd, "completed", fmt.Sprintf("Job executed successfully at %s", pastTime.Format("15:04:05")))
                }
        }
}