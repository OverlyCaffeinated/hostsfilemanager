// hostfilemanager is a small CLI tool that reads a hosts.json file and
// creates/updates a managed section in the system hosts file (/etc/hosts).
//
// The managed section is delimited by marker comments so that:
//   * the file owner can clearly see which records are managed by this app,
//   * the app knows exactly where to insert/update the records it owns.
//
// Everything outside the managed section is left untouched.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

const (
	// hostsPath is the location of the system hosts file on Linux/macOS.
	hostsPath = "/etc/hosts"

	// configPath is the default location of the input file.
	configPath = "hosts.json"

	// beginMarker / endMarker delimit the managed section inside /etc/hosts.
	beginMarker = "# BEGIN hostfilemanager managed entries"
	endMarker   = "# END hostfilemanager managed entries"

	// headerComment explains the purpose of the section to a human reader.
	headerComment = "# Managed by hostfilemanager - entries between these markers are managed automatically"
)

// HostEntry is a single hosts-file record: one IP mapped to one or more hostnames.
type HostEntry struct {
	IP       string   `json:"ip"`
	Hostnames []string `json:"hostnames"`
	Comment  string   `json:"comment,omitempty"`
}

func main() {
	// Allow an alternate config path via the first argument.
	configPath := defaultConfigPath
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	if err := run(configPath); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

const defaultConfigPath = configPath

// run loads the config, builds the managed block, merges it into the hosts
// file, and writes the result back.
func run(configPath string) error {
	entries, err := loadConfig(configPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	hostsLines, err := readLines(hostsPath)
	if err != nil {
		return fmt.Errorf("reading hosts file: %w", err)
	}

	managedBlock := buildManagedBlock(entries)
	merged := mergeManagedSection(hostsLines, managedBlock)

	if err := writeLines(hostsPath, merged); err != nil {
		return fmt.Errorf("writing hosts file: %w", err)
	}

	fmt.Printf("hosts file updated: %d managed entry(ies) written to %s\n", len(entries), hostsPath)
	return nil
}

// loadConfig reads and parses the hosts.json file.
func loadConfig(path string) ([]HostEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var entries []HostEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("parsing JSON: %w", err)
	}

	if entries == nil {
		entries = []HostEntry{}
	}

	return entries, nil
}

// readLines reads the hosts file into a slice of lines (without trailing newlines).
func readLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	// Increase the buffer so very long lines (rare in hosts files) still scan.
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}

// buildManagedBlock produces the full set of lines that make up the managed
// section, including the marker comments and header.
func buildManagedBlock(entries []HostEntry) []string {
	var block []string

	block = append(block, beginMarker)
	block = append(block, headerComment)

	for _, entry := range entries {
		if strings.TrimSpace(entry.IP) == "" || len(entry.Hostnames) == 0 {
			continue
		}

		line := entry.IP + "\t" + strings.Join(entry.Hostnames, "\t")
		if entry.Comment != "" {
			line += "  # " + entry.Comment
		}
		block = append(block, line)
	}

	block = append(block, endMarker)
	return block
}

// mergeManagedSection replaces any existing managed section between the
// begin/end markers with the new block. If no section exists, the block is
// appended to the end of the file. Everything outside the section is preserved.
func mergeManagedSection(hostsLines []string, managedBlock []string) []string {
	beginIdx := -1
	endIdx := -1

	for i, line := range hostsLines {
		trimmed := strings.TrimSpace(line)
		switch {
		case trimmed == beginMarker:
			beginIdx = i
		case trimmed == endMarker && beginIdx >= 0 && endIdx < 0:
			endIdx = i
		}
	}

	result := make([]string, 0, len(hostsLines)+len(managedBlock))

	if beginIdx >= 0 && endIdx >= 0 {
		// Keep everything before the section, insert the new block, keep everything after.
		result = append(result, hostsLines[:beginIdx]...)
		result = append(result, managedBlock...)
		result = append(result, hostsLines[endIdx+1:]...)
	} else {
		// No existing section: append to the end, separated by a blank line.
		result = append(result, hostsLines...)
		if len(result) > 0 && strings.TrimSpace(result[len(result)-1]) != "" {
			result = append(result, "")
		}
		result = append(result, "")
		result = append(result, managedBlock...)
	}

	return result
}

// writeLines writes the lines back to the hosts file, each followed by a newline.
func writeLines(path string, lines []string) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	for _, line := range lines {
		if _, err := fmt.Fprintln(w, line); err != nil {
			return err
		}
	}
	return w.Flush()
}
