package main

import (
        "bufio"
        "fmt"
        "os"
        "os/exec"
        "regexp"
        "sort"
        "strings"
        "time"
)

// LogEntry represents a single log entry for a cron job
type LogEntry struct {
        Timestamp time.Time
        Status    string
        Message   string
}

// GetLastRunTime attempts to find the last execution time of a command
func GetLastRunTime(command string) time.Time {
        // First check our custom history
        if lastRun := getLastRunFromHistory(command); !lastRun.IsZero() {
                return lastRun
        }

        // Try different approaches to find the last run time
        if lastRun := checkSystemdJournal(command); !lastRun.IsZero() {
                return lastRun
        }

        if lastRun := checkSyslog(command); !lastRun.IsZero() {
                return lastRun
        }

        if lastRun := checkCronLog(command); !lastRun.IsZero() {
                return lastRun
        }

        return time.Time{} // Return zero time if not found
}

// GetJobHistory retrieves the execution history for a specific command
func GetJobHistory(command string) []LogEntry {
        var entries []LogEntry

        // First get our custom tracked history
        if customEntries := getHistoryFromFile(command); len(customEntries) > 0 {
                entries = append(entries, customEntries...)
        }

        // Try systemd journal
        if journalEntries := getSystemdJournalEntries(command); len(journalEntries) > 0 {
                entries = append(entries, journalEntries...)
        }

        // Try syslog
        if syslogEntries := getSyslogEntries(command); len(syslogEntries) > 0 {
                entries = append(entries, syslogEntries...)
        }

        // Try cron.log
        if cronEntries := getCronLogEntries(command); len(cronEntries) > 0 {
                entries = append(entries, cronEntries...)
        }

        // Sort by timestamp (newest first)
        sort.Slice(entries, func(i, j int) bool {
                return entries[i].Timestamp.After(entries[j].Timestamp)
        })

        // Remove duplicates and limit to last 50 entries
        seen := make(map[string]bool)
        var unique []LogEntry
        for _, entry := range entries {
                key := entry.Timestamp.Format("2006-01-02 15:04:05") + entry.Message
                if !seen[key] && len(unique) < 50 {
                        seen[key] = true
                        unique = append(unique, entry)
                }
        }

        return unique
}

// checkSystemdJournal checks systemd journal for the last execution
func checkSystemdJournal(command string) time.Time {
        // Extract the main command (first word)
        mainCmd := strings.Fields(command)[0]
        if strings.HasPrefix(mainCmd, "/") {
                parts := strings.Split(mainCmd, "/")
                mainCmd = parts[len(parts)-1]
        }

        cmd := exec.Command("journalctl", "--user", "-u", "cron", "--since", "1 month ago", "--grep", mainCmd, "-n", "1", "--output", "short-iso")
        output, err := cmd.Output()
        if err != nil {
                return time.Time{}
        }

        return parseTimestampFromJournal(string(output))
}

// getSystemdJournalEntries retrieves systemd journal entries for a command
func getSystemdJournalEntries(command string) []LogEntry {
        mainCmd := strings.Fields(command)[0]
        if strings.HasPrefix(mainCmd, "/") {
                parts := strings.Split(mainCmd, "/")
                mainCmd = parts[len(parts)-1]
        }

        cmd := exec.Command("journalctl", "--user", "-u", "cron", "--since", "1 month ago", "--grep", mainCmd, "-n", "100", "--output", "short-iso")
        output, err := cmd.Output()
        if err != nil {
                return nil
        }

        return parseJournalEntries(string(output))
}

// checkSyslog checks /var/log/syslog for cron entries
func checkSyslog(command string) time.Time {
        file, err := os.Open("/var/log/syslog")
        if err != nil {
                return time.Time{}
        }
        defer file.Close()

        mainCmd := strings.Fields(command)[0]
        if strings.HasPrefix(mainCmd, "/") {
                parts := strings.Split(mainCmd, "/")
                mainCmd = parts[len(parts)-1]
        }

        scanner := bufio.NewScanner(file)
        var lastTime time.Time

        cronRegex := regexp.MustCompile(`(\w{3}\s+\d{1,2}\s+\d{2}:\d{2}:\d{2}).*CRON.*` + regexp.QuoteMeta(mainCmd))

        for scanner.Scan() {
                line := scanner.Text()
                if matches := cronRegex.FindStringSubmatch(line); matches != nil {
                        if t := parseSyslogTimestamp(matches[1]); !t.IsZero() && t.After(lastTime) {
                                lastTime = t
                        }
                }
        }

        return lastTime
}

// getSyslogEntries retrieves syslog entries for a command
func getSyslogEntries(command string) []LogEntry {
        var entries []LogEntry
        
        files := []string{"/var/log/syslog", "/var/log/syslog.1"}
        
        mainCmd := strings.Fields(command)[0]
        if strings.HasPrefix(mainCmd, "/") {
                parts := strings.Split(mainCmd, "/")
                mainCmd = parts[len(parts)-1]
        }

        cronRegex := regexp.MustCompile(`(\w{3}\s+\d{1,2}\s+\d{2}:\d{2}:\d{2}).*CRON.*(.*)` + regexp.QuoteMeta(mainCmd) + `(.*)`)

        for _, filename := range files {
                file, err := os.Open(filename)
                if err != nil {
                        continue
                }

                scanner := bufio.NewScanner(file)
                for scanner.Scan() {
                        line := scanner.Text()
                        if matches := cronRegex.FindStringSubmatch(line); matches != nil {
                                if t := parseSyslogTimestamp(matches[1]); !t.IsZero() {
                                        entry := LogEntry{
                                                Timestamp: t,
                                                Status:    "executed",
                                                Message:   strings.TrimSpace(matches[2] + matches[3]),
                                        }
                                        entries = append(entries, entry)
                                }
                        }
                }
                file.Close()
        }

        return entries
}

// checkCronLog checks /var/log/cron for cron entries
func checkCronLog(command string) time.Time {
        file, err := os.Open("/var/log/cron")
        if err != nil {
                return time.Time{}
        }
        defer file.Close()

        mainCmd := strings.Fields(command)[0]
        if strings.HasPrefix(mainCmd, "/") {
                parts := strings.Split(mainCmd, "/")
                mainCmd = parts[len(parts)-1]
        }

        scanner := bufio.NewScanner(file)
        var lastTime time.Time

        cronRegex := regexp.MustCompile(`(\w{3}\s+\d{1,2}\s+\d{2}:\d{2}:\d{2}).*` + regexp.QuoteMeta(mainCmd))

        for scanner.Scan() {
                line := scanner.Text()
                if matches := cronRegex.FindStringSubmatch(line); matches != nil {
                        if t := parseSyslogTimestamp(matches[1]); !t.IsZero() && t.After(lastTime) {
                                lastTime = t
                        }
                }
        }

        return lastTime
}

// getCronLogEntries retrieves cron log entries for a command
func getCronLogEntries(command string) []LogEntry {
        var entries []LogEntry
        
        files := []string{"/var/log/cron", "/var/log/cron.1"}
        
        mainCmd := strings.Fields(command)[0]
        if strings.HasPrefix(mainCmd, "/") {
                parts := strings.Split(mainCmd, "/")
                mainCmd = parts[len(parts)-1]
        }

        cronRegex := regexp.MustCompile(`(\w{3}\s+\d{1,2}\s+\d{2}:\d{2}:\d{2}).*(.*)` + regexp.QuoteMeta(mainCmd) + `(.*)`)

        for _, filename := range files {
                file, err := os.Open(filename)
                if err != nil {
                        continue
                }

                scanner := bufio.NewScanner(file)
                for scanner.Scan() {
                        line := scanner.Text()
                        if matches := cronRegex.FindStringSubmatch(line); matches != nil {
                                if t := parseSyslogTimestamp(matches[1]); !t.IsZero() {
                                        entry := LogEntry{
                                                Timestamp: t,
                                                Status:    "executed",
                                                Message:   strings.TrimSpace(matches[2] + matches[3]),
                                        }
                                        entries = append(entries, entry)
                                }
                        }
                }
                file.Close()
        }

        return entries
}

// parseTimestampFromJournal parses timestamp from systemd journal output
func parseTimestampFromJournal(output string) time.Time {
        lines := strings.Split(strings.TrimSpace(output), "\n")
        if len(lines) == 0 {
                return time.Time{}
        }

        // Journal format: 2023-08-12T10:30:15+0000
        timestampRegex := regexp.MustCompile(`(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}[\+\-]\d{4})`)
        if matches := timestampRegex.FindStringSubmatch(lines[0]); matches != nil {
                if t, err := time.Parse("2006-01-02T15:04:05-0700", matches[1]); err == nil {
                        return t
                }
        }

        return time.Time{}
}

// parseJournalEntries parses multiple journal entries
func parseJournalEntries(output string) []LogEntry {
        var entries []LogEntry
        lines := strings.Split(strings.TrimSpace(output), "\n")

        timestampRegex := regexp.MustCompile(`(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}[\+\-]\d{4})\s+\S+\s+(.*)`)

        for _, line := range lines {
                if matches := timestampRegex.FindStringSubmatch(line); matches != nil {
                        if t, err := time.Parse("2006-01-02T15:04:05-0700", matches[1]); err == nil {
                                entry := LogEntry{
                                        Timestamp: t,
                                        Status:    "executed",
                                        Message:   matches[2],
                                }
                                entries = append(entries, entry)
                        }
                }
        }

        return entries
}

// parseSyslogTimestamp parses syslog timestamp format
func parseSyslogTimestamp(timestamp string) time.Time {
        // Format: Aug 12 10:30:15
        currentYear := time.Now().Year()
        timestampWithYear := fmt.Sprintf("%d %s", currentYear, timestamp)
        
        if t, err := time.Parse("2006 Jan 2 15:04:05", timestampWithYear); err == nil {
                // If the parsed time is in the future, it's probably from last year
                if t.After(time.Now()) {
                        t = t.AddDate(-1, 0, 0)
                }
                return t
        }

        return time.Time{}
}
