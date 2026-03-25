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

const (
	VERSION  = "0.1.0"
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
	baseDir      = filepath.Join(home, ".local/share/kepfi") // Main app data directory
	trashDir     = filepath.Join(baseDir, "trash")           // Directory where files are moved
	metadataFile = filepath.Join(baseDir, "metadata.json")   // JSON file tracking original paths

	// Command-line flags
	showVersion = flag.Bool("v", false, "Display version")
	restore     = flag.String("r", "", "Restore a file/folder by name")
	temp        = flag.Bool("t", false, "Move to /tmp/")
	list        = flag.Bool("l", false, "List all files/folders in kepfi trash")
	listRemove  = flag.String("lr", "", "Delete a specific file/folder in kepfi trash")
	removeAll   = flag.Bool("rm", false, "Delete all files/folders from kepfi trash")
	force       = flag.Bool("f", false, "Force action (no confirmation)")
	schedule    = flag.String("at", "", "Schedule a one-time purge at HH:MM")
)

// FileRecord defines the structure for metadata storage
type FileRecord struct {
	FileName     string    `json:"file_name"`
	OriginalPath string    `json:"original_path"`
	DeletedAt    time.Time `json:"deleted_at"`
	IsTemp       bool      `json:"is_temp"`
}

// init ensures the trash directory exists before the app runs
func init() {
	_ = os.MkdirAll(trashDir, 0755)
}

// ========================= HELPER FUNCTIONS =========================

// loadRecords reads and decodes the metadata JSON file
func loadRecords() []FileRecord {
	var r []FileRecord
	data, err := os.ReadFile(metadataFile)
	if err != nil {
		return []FileRecord{} // Return empty list if file doesn't exist
	}
	json.Unmarshal(data, &r)
	return r
}

// writeRecords encodes and saves the records to the metadata JSON file
func writeRecords(records []FileRecord) {
	data, _ := json.MarshalIndent(records, "", "  ")
	_ = os.WriteFile(metadataFile, data, 0644)
}

// saveRecord appends a new deletion entry to the metadata
func saveRecord(record FileRecord) {
	records := loadRecords()
	records = append(records, record)
	writeRecords(records)
}

// truncate shortens strings that are too long for the terminal table view
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

// formatSize converts bytes into human-readable strings (MB, GB, etc.)
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

// getDirSize calculates the total size of a directory recursively
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

// getTerminalWidth fetches the current terminal size for dynamic UI scaling
func getTerminalWidth() int {
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()
	if err != nil {
		return 80 // Default width if command fails
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
	flag.Parse() // Parse all user-provided flags

	args := flag.Args() // Get non-flag arguments (filenames)

	if *showVersion {
		fmt.Printf("kepfi %s\n", VERSION)
		fmt.Printf("Made by Knuspii, (M)\n")
		return
	}

	fmt.Printf("%s[kepfi]%s\n", CYAN, RC)

	// Route to the correct function based on flags
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

	if *schedule != "" {
		schedulePurge(*schedule)
		return
	}

	// Handle standard file trashing
	if len(args) > 0 {
		for _, path := range args {
			absPath, _ := filepath.Abs(path)

			// SAFETY CHECK: Prevent accidental deletion of system or root folders
			if absPath == "/" || absPath == home || path == "." || path == ".." {
				fmt.Printf("%sError: Path '%s' is too dangerous to move%s\n", RED, path, RC)
				continue
			}

			moveToTrash(path, *temp)
		}
	} else {
		// Display basic usage if no arguments provided
		fmt.Printf("Usage: kepfi <filename>\n")
		fmt.Printf("Use 'kepfi -h' for help\n")
	}
}

// listRecords displays a formatted table of all items currently in the trash
func listRecords() {
	records := loadRecords()
	if len(records) == 0 {
		fmt.Printf("kepfi trash list is empty\n")
		return
	}

	// UI Layout Variables
	nameW := 18 // Width for filename column
	typeW := 5  // Width for DIR/FILE indicator
	dateW := 16 // Width for timestamp
	termWidth := getTerminalWidth()

	fmt.Printf("%skepfi trash list:%s\n\n", CYAN, RC)

	// Print Table Header
	fmt.Printf("%s%-*s %-*s %-*s %s%s\n",
		BOLD,
		nameW, "FILENAME:",
		typeW, "TYPE:",
		dateW, "DELETED AT:",
		"ORIGINAL PATH:", RC)

	fmt.Printf("%s%s%s\n", GRAY, strings.Repeat("-", termWidth), RC)

	// Iterate and print each record
	for _, r := range records {
		displayName := truncate(r.FileName, nameW)

		// Determine if the item is a folder or a file
		fileType := "FILE"
		info, err := os.Stat(filepath.Join(trashDir, r.FileName))
		if err == nil && info.IsDir() {
			fileType = "DIR"
		}

		// Dynamically calculate space for path column
		pathLimit := termWidth - nameW - typeW - dateW - 3
		if pathLimit < 10 {
			pathLimit = 10
		}
		displayPath := truncate(r.OriginalPath, pathLimit)

		fmt.Printf("%-*s %-*s %-*s %s%s%s\n",
			nameW, displayName,
			typeW, fileType,
			dateW, r.DeletedAt.Format("2006-01-02 15:04"),
			GRAY, displayPath, RC)
	}

	fmt.Printf("%s%s%s\n", GRAY, strings.Repeat("-", termWidth), RC)
	totalSize, _ := getDirSize(trashDir)
	fmt.Printf("kepfi trash total size: %s%s%s\n", YELLOW, formatSize(totalSize), RC)
}

// restoreFile moves a trashed file back to its original location
func restoreFile(name string) {
	records := loadRecords()
	var updated []FileRecord
	found := false

	for _, r := range records {
		if r.FileName == name && !found {
			src := filepath.Join(trashDir, r.FileName)
			destDir := filepath.Dir(r.OriginalPath)

			// Ensure the original folder still exists (or recreate it)
			_ = os.MkdirAll(destDir, 0755)

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
		fmt.Printf("%sError: '%s' not found in kepfi trash%s\n", RED, name, RC)
		return
	}
	writeRecords(updated) // Save metadata without the restored file
}

// removeSpecific permanently deletes a single item from the trash
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

			err = os.RemoveAll(targetPath) // Permanent deletion
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
		fmt.Printf("%sError: '%s' not found in kepfi trash%s\n", RED, name, RC)
		return
	}

	writeRecords(updated)
	fmt.Printf("%s[DONE]%s Permanently removed '%s' (%s cleared)\n", GREEN, RC, name, formatSize(size))
}

// schedulePurge launches a background process that cleans the trash at a certain time
func schedulePurge(targetTime string) {
	parts := strings.Split(targetTime, ":")
	if len(parts) != 2 {
		fmt.Printf("%sError: Use HH:MM format%s\n", RED, RC)
		return
	}
	hour, _ := strconv.Atoi(parts[0])
	min, _ := strconv.Atoi(parts[1])

	now := time.Now()
	target := time.Date(now.Year(), now.Month(), now.Day(), hour, min, 0, 0, now.Location())

	// If time is already past today, set for tomorrow
	if target.Before(now) {
		target = target.Add(24 * time.Hour)
	}

	delay := int(time.Until(target).Seconds())
	exe, _ := os.Executable() // Absolute path to the kepfi binary

	// Creates a shell command that sleeps and then triggers a forced purge
	cmdStr := fmt.Sprintf("sleep %d && %s -rm -f > /dev/null 2>&1", delay, exe)
	cmd := exec.Command("sh", "-c", cmdStr)

	err := cmd.Start() // Run in background
	if err != nil {
		fmt.Printf("%sError: Failed to background process: %v%s\n", RED, err, RC)
		return
	}

	fmt.Printf("%s[SCHEDULED]%s\n", CYAN, RC)
	fmt.Printf("  Target: %s (%v from now)\n", target.Format("15:04"), time.Until(target).Round(time.Second))
	fmt.Printf("  Action: %sFull purge of kepfi trash%s\n", GREEN, RC)
	fmt.Printf("  Status: Running in background (PID: %s%d%s)\n", YELLOW, cmd.Process.Pid, RC)
	fmt.Printf("  %sNote: Process will die if you shutdown before the target time%s\n", GRAY, RC)

	cmd.Process.Release() // Detach from current process so it survives terminal closing
}

// moveToTrash handles the renaming of a file from its original path to the trash folder
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

	// Handle temporary deletion (/tmp/)
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

	// Move to standard kepfi trash
	destPath = filepath.Join(trashDir, fileName)
	err = os.Rename(absPath, destPath)
	if err != nil {
		fmt.Printf("%sError: Failed to move to kepfi trash: %v%s\n", RED, err, RC)
		return
	}

	// Save the record so it can be restored or listed later
	saveRecord(FileRecord{
		FileName:     fileName,
		OriginalPath: absPath,
		DeletedAt:    time.Now(),
		IsTemp:       false,
	})

	fmt.Printf("%s[TRASH]%s '%s' moved to trash\n", GREEN, RC, fileName)
}

// purgeEverything completely wipes the trash folder and metadata file
func purgeEverything() {
	records := loadRecords()
	size, _ := getDirSize(trashDir)

	// Check if trash is already empty
	if size == 0 && len(records) == 0 {
		fmt.Printf("kepfi trash is already empty\n")
		return
	}

	// Require confirmation unless -f flag is used
	if !*force {
		fmt.Printf("%s%sWARNING: This will permanently delete everything in kepfi trash!%s\n", BOLD, RED, RC)
		fmt.Printf("Confirm action? (y/N): ")
		var response string
		fmt.Scanln(&response)

		if strings.ToLower(response) != "y" {
			fmt.Printf("Action cancelled\n")
			return
		}
	}

	_ = os.Remove(metadataFile)      // Remove metadata
	errDir := os.RemoveAll(trashDir) // Remove files
	_ = os.MkdirAll(trashDir, 0755)  // Recreate empty trash folder

	if errDir != nil {
		fmt.Printf("%sError: Failed to clear trash folder%s\n", RED, RC)
		os.Exit(1)
	}

	fmt.Printf("%s[DONE]%s kepfi trash purged. %s%s%s cleared\n", GREEN, RC, BOLD, formatSize(size), RC)
}
