package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// ANSI Color Codes
const (
	VERSION  = "0.0.1"
	cacheDir = "/tmp"
	RC       = "\033[0m"
	BOLD     = "\033[1m"
	RED      = "\033[31m"
	GREEN    = "\033[32m"
	YELLOW   = "\033[33m"
	CYAN     = "\033[36m"
	GRAY     = "\033[90m"
)

var (
	home, _      = os.UserHomeDir()
	baseDir      = filepath.Join(home, ".local/share/kepfi")
	trashDir     = filepath.Join(baseDir, "trash")
	metadataFile = filepath.Join(baseDir, "metadata.json")
)

type FileRecord struct {
	FileName     string    `json:"file_name"`
	OriginalPath string    `json:"original_path"`
	DeletedAt    time.Time `json:"deleted_at"`
	IsTemp       bool      `json:"is_temp"`
}

func init() {
	_ = os.MkdirAll(trashDir, 0755)
}

// ========================= HELPER FUNCTIONS =========================

func loadRecords() []FileRecord {
	var r []FileRecord
	data, err := os.ReadFile(metadataFile)
	if err != nil {
		return []FileRecord{}
	}
	json.Unmarshal(data, &r)
	return r
}

func writeRecords(records []FileRecord) {
	data, _ := json.MarshalIndent(records, "", "  ")
	_ = os.WriteFile(metadataFile, data, 0644)
}

func saveRecord(record FileRecord) {
	records := loadRecords()
	records = append(records, record)
	writeRecords(records)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func getDirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}

func getTerminalWidth() int {
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()
	if err != nil {
		return 80
	}
	parts := strings.Split(strings.TrimSpace(string(out)), " ")
	if len(parts) < 2 {
		return 80
	}
	width, _ := strconv.Atoi(parts[1])
	return width
}

// ========================= MAIN FUNCTIONS =========================

func main() {
	showVersion := flag.Bool("v", false, "Display version")
	restore := flag.String("r", "", "Restore a file/folder by name")
	temp := flag.Bool("t", false, "Move to /tmp/")
	list := flag.Bool("l", false, "List all files/folders in kepfi trash")
	listRemove := flag.String("lr", "", "Delete a specific file/folder in kepfi trash")
	removeAll := flag.Bool("rm", false, "Delete all files/folders from kepfi trash")
	flag.Parse()

	args := flag.Args()

	if *showVersion {
		fmt.Printf("kepfi %s\n", VERSION)
		fmt.Printf("Made by Knuspii, (M)\n")
		return
	}

	fmt.Printf("%s[kepfi]%s\n", CYAN, RC)

	if *removeAll {
		purgeEverything()
		return
	}

	if *list {
		listRecords()
		return
	}

	if *restore != "" {
		restoreFile(*restore)
		return
	}

	if *listRemove != "" {
		removeSpecific(*listRemove)
		return
	}

	if len(args) > 0 {
		for _, path := range args {
			if path == "." || path == "/" || path == "*" {
				fmt.Printf("%sError: Path '%s' is too dangerous to move%s\n", RED, path, RC)
				continue
			}
			moveToTrash(path, *temp)
		}
	} else {
		fmt.Printf("Usage: kepfi <filename>\n")
		fmt.Printf("Use 'kepfi -h' for help\n")
	}
}

func listRecords() {
	records := loadRecords()
	if len(records) == 0 {
		fmt.Printf("Trash list is empty\n")
		return
	}

	// Configuration variables
	nameW := 18 // Width for Filename
	dateW := 16 // Width for Date (YYYY-MM-DD HH:MM)
	termWidth := getTerminalWidth()

	fmt.Printf("%skepfi trash list:%s\n\n", CYAN, RC)

	// Header - using %-*s (left-aligned with variable width)
	fmt.Printf("%s%-*s %-*s %s%s\n",
		BOLD,
		nameW, "FILENAME:",
		dateW, "DELETED AT:",
		"ORIGINAL PATH:", RC)

	// Dynamic separator line
	fmt.Printf("%s%s%s\n", GRAY, strings.Repeat("-", termWidth), RC)

	for _, r := range records {
		displayName := truncate(r.FileName, nameW)

		// Calculate path limit dynamically based on your two variables
		pathLimit := termWidth - nameW - dateW - 2
		if pathLimit < 10 {
			pathLimit = 10
		}
		displayPath := truncate(r.OriginalPath, pathLimit)

		// Row - using the variables for consistent alignment
		fmt.Printf("%-*s %-*s %s%s%s\n",
			nameW, displayName,
			dateW, r.DeletedAt.Format("2006-01-02 15:04"),
			GRAY, displayPath, RC)
	}

	fmt.Printf("%s%s%s\n", GRAY, strings.Repeat("-", termWidth), RC)
	totalSize, _ := getDirSize(trashDir)
	fmt.Printf("Total trash size: %s%s%s\n", YELLOW, formatSize(totalSize), RC)
}

func restoreFile(name string) {
	records := loadRecords()
	var updated []FileRecord
	found := false

	for _, r := range records {
		if r.FileName == name && !found {
			src := filepath.Join(trashDir, r.FileName)
			err := os.Rename(src, r.OriginalPath)
			if err != nil {
				fmt.Printf("%sError: Could not restore file: %v%s\n", RED, err, RC)
				os.Exit(1)
			}
			fmt.Printf("%s[RESTORED]%s %s to %s\n", GREEN, RC, r.FileName, r.OriginalPath)
			found = true
			continue
		}
		updated = append(updated, r)
	}

	if !found {
		fmt.Printf("%sError: '%s' not found in trash%s\n", RED, name, RC)
		return
	}
	writeRecords(updated)
}

func removeSpecific(name string) {
	records := loadRecords()
	var updated []FileRecord
	found := false
	var size int64

	for _, r := range records {
		if r.FileName == name && !found {
			targetPath := filepath.Join(trashDir, r.FileName)
			info, err := os.Stat(targetPath)
			if err == nil {
				size = info.Size()
			}

			err = os.RemoveAll(targetPath)
			if err != nil {
				fmt.Printf("%sError: Could not delete '%s': %v%s\n", RED, name, err, RC)
				os.Exit(1)
			}
			found = true
			continue
		}
		updated = append(updated, r)
	}

	if !found {
		fmt.Printf("%sError: '%s' not found in trash%s\n", RED, name, RC)
		return
	}

	writeRecords(updated)
	fmt.Printf("%s[DONE]%s Permanently removed '%s' (%s cleared)\n", GREEN, RC, name, formatSize(size))
}

func purgeEverything() {
	size, _ := getDirSize(trashDir)
	if size == 0 {
		fmt.Printf("Trash is already empty\n")
		return
	}

	fmt.Printf("%s%sWARNING: This will permanently delete everything in trash!%s\n", BOLD, RED, RC)
	fmt.Printf("Confirm action? (y/N): ")
	var response string
	fmt.Scanln(&response)

	if strings.ToLower(response) != "y" {
		fmt.Printf("Action cancelled\n")
		return
	}

	errDir := os.RemoveAll(trashDir)
	errMeta := os.Remove(metadataFile)
	_ = os.MkdirAll(trashDir, 0755)

	if errDir != nil || errMeta != nil {
		fmt.Printf("%sError: Failed to clear trash folders or metadata%s\n", RED, RC)
		os.Exit(1)
	}

	fmt.Printf("%s[DONE]%s Trash purged. %s%s%s cleared.\n", GREEN, RC, BOLD, formatSize(size), RC)
}

func moveToTrash(target string, isTemp bool) {
	absPath, err := filepath.Abs(target)
	if err != nil {
		fmt.Printf("%sError: Path resolution failed for '%s'%s\n", RED, target, RC)
		return
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		fmt.Printf("%sError: '%s' does not exist%s\n", RED, target, RC)
		return
	}

	fileName := filepath.Base(absPath)
	var destPath string

	if isTemp {
		destPath = filepath.Join(cacheDir, fileName)
		err = os.Rename(absPath, destPath)
		if err != nil {
			fmt.Printf("%sError: Failed to move to temp: %v%s\n", RED, err, RC)
			return
		}
		fmt.Printf("%s[TEMP]%s '%s' moved to /tmp/\n", YELLOW, RC, fileName)
		return
	}

	destPath = filepath.Join(trashDir, fileName)
	err = os.Rename(absPath, destPath)
	if err != nil {
		fmt.Printf("%sError: Failed to move to trash: %v%s\n", RED, err, RC)
		return
	}

	saveRecord(FileRecord{
		FileName:     fileName,
		OriginalPath: absPath,
		DeletedAt:    time.Now(),
		IsTemp:       false,
	})

	fmt.Printf("%s[TRASH]%s '%s' moved to trash\n", GREEN, RC, fileName)
}
