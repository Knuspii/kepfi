package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const USAGE = `A smart alternative to rm with a recovery bin and storage tracking.

Project URL: https://github.com/Knuspii/kepfi
License: GPL-3.0
Author: Knuspii, (M)

Usage: kepfi [OPTION]

Options:
  -l,  --list                   Shows a detailed table of kepfi trashed items
  -r,  --restore <FILE>         Restores a file/folder back to its original location
  -t,  --temp <FILE>            Move a file/folder to /tmp/
  -ps, --purge-specific <FILE>  Purge specific file/folder in kepfi trash
  -pa, --purge-all              Purge all files/folders in kepfi trash
  -f,  --force                  Force action (no confirmation)
  -at, --at-time <HH:MM>        Schedule a one-time purge at a specific time
  -v,  --version                Displays version and infos

Examples:
kepfi file.txt        Move file.txt to kepfi trash
kepfi -r file.txt     Restore file.txt to its original path
kepfi -at 22:30       Schedule a background purge for 22:30

`

const (
	VERSION  = "0.2.1"
	CACHEDIR = "/tmp"
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

	optVersion   bool
	optRestore   string
	optTemp      bool
	optList      bool
	optPurgeSpec string
	optPurgeAll  bool
	optForce     bool
	optAt        string
	fileArgs     []string
)

// FileRecord defines the structure for metadata storage
type FileRecord struct {
	FileName     string    `json:"file_name"`
	OriginalPath string    `json:"original_path"`
	DeletedAt    time.Time `json:"deleted_at"`
}

// init ensures the trash directory exists before the app runs
func init() {
	err := os.MkdirAll(trashDir, 0755)
	if err != nil {
		fmt.Printf("%sError: Could not init: %v%s\n", RED, err, RC)
	}
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
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
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

// getUniquePath ensures no overwrite by appending (1), (2), etc.
func getUniquePath(dir, fileName string) string {
	ext := filepath.Ext(fileName)
	name := strings.TrimSuffix(fileName, ext)

	newPath := filepath.Join(dir, fileName)
	counter := 1

	for {
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			return newPath
		}

		newName := fmt.Sprintf("%s(%d)%s", name, counter, ext)
		newPath = filepath.Join(dir, newName)
		counter++
	}
}

// ========================= MAIN FUNCTIONS =========================

func main() {
	args := os.Args[1:]

	// Custom Argument Parser
	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch arg {
		case "-v", "-V", "--version":
			optVersion = true
		case "-l", "--list":
			optList = true
		case "-pa", "--purge-all":
			optPurgeAll = true
		case "-f", "--force":
			optForce = true
		case "-t", "--temp":
			optTemp = true
		case "-r", "--restore":
			if i+1 < len(args) {
				optRestore = args[i+1]
				i++
			}
		case "-ps", "--purge-specific":
			if i+1 < len(args) {
				optPurgeSpec = args[i+1]
				i++
			}
		case "-at", "--at-time":
			if i+1 >= len(args) || strings.HasPrefix(args[i+1], "-") {
				fmt.Printf("%sError: -at requires a time (HH:MM)%s\n", RED, RC)
				os.Exit(1)
			} else {
				optAt = args[i+1]
				i++
			}
		case "-h", "--help":
			fmt.Print(USAGE)
			return
		default:
			if !strings.HasPrefix(arg, "-") {
				fileArgs = append(fileArgs, arg)
			} else {
				fmt.Printf("%sUnknown flag: %s%s\n", RED, arg, RC)
				fmt.Printf("Use 'kepfi -h' for help\n")
				os.Exit(1)
			}
		}
	}

	// Execution Logic
	if optVersion {
		fmt.Printf("kepfi %s\nMade by Knuspii, (M)\n", VERSION)
		return
	}

	fmt.Printf("%s[kepfi]%s\n", CYAN, RC)
	// Check if running as root/sudo
	if os.Geteuid() == 0 {
		fmt.Printf("%s[NOTICE]%s Running with root privileges (sudo). ", YELLOW, RC)
		fmt.Printf("%sYour trash will be located in %s%s\n", GRAY, trashDir, RC)
	}

	if optList {
		listRecords()
		return
	}

	if optRestore != "" {
		restoreFile(optRestore)
		return
	}

	if optPurgeSpec != "" {
		removeSpecific(optPurgeSpec)
		return
	}

	if optPurgeAll {
		purgeEverything()
		return
	}

	if optAt != "" {
		schedulePurge(optAt)
		return
	}

	// Handle standard file trashing
	if len(fileArgs) > 0 {
		for _, path := range fileArgs {
			absPath, _ := filepath.Abs(path)

			// SAFETY CHECK
			if absPath == "/" || absPath == trashDir || absPath == home || path == "." || path == ".." {
				fmt.Printf("%sError: Path '%s' is too dangerous to move%s\n", RED, path, RC)
				continue
			}
			moveToTrash(path, optTemp)
		}
	} else {
		fmt.Printf("%sError: You need to provide a filename%s\n", RED, RC)
		fmt.Printf("Use 'kepfi -h' for help\n")
		os.Exit(1)
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
	nameW := 25 // Width for filename column
	typeW := 5  // Width for DIR/FILE indicator
	sizeW := 8  // Width for Size column
	dateW := 16 // Width for timestamp
	termWidth := getTerminalWidth()

	fmt.Printf("kepfi trash list:\n\n")

	// Print Table Header
	fmt.Printf("%s%-*s %-*s %-*s %-*s %s%s\n",
		BOLD,
		nameW, "FILENAME:",
		typeW, "TYPE:",
		sizeW, "SIZE:",
		dateW, "DELETED AT:",
		"ORIGINAL PATH:", RC)

	fmt.Printf("%s%s%s\n", GRAY, strings.Repeat("-", termWidth), RC)

	// Iterate and print each record
	for _, r := range records {
		displayName := truncate(r.FileName, nameW)
		itemPath := filepath.Join(trashDir, r.FileName)

		// Determine if the item is a folder or a file
		fileType := "FILE"
		info, err := os.Stat(itemPath)
		if err == nil && info.IsDir() {
			fileType = "DIR"
		}

		// Calculate size
		sizeBytes, _ := getDirSize(itemPath)
		displaySize := formatSize(sizeBytes)

		// Dynamically calculate space for path column
		// We subtract the widths of other columns and the spaces between them
		pathLimit := termWidth - nameW - typeW - sizeW - dateW - 4
		if pathLimit < 10 {
			pathLimit = 10
		}
		displayPath := truncate(r.OriginalPath, pathLimit)

		fmt.Printf("%-*s %-*s %-*s %-*s %s%s%s\n",
			nameW, displayName,
			typeW, fileType,
			sizeW, displaySize,
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

			// check if destination already exists
			destPath := r.OriginalPath
			if _, err := os.Stat(destPath); err == nil {
				// file exists -> generate new safe name
				destPath = getUniquePath(destDir, filepath.Base(destPath))
			}

			// now restore safely
			err := os.Rename(src, destPath)
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
		// Only target the specific file and only the first match
		if r.FileName == name && !found {
			targetPath := filepath.Join(trashDir, r.FileName)

			// FIX: Use getDirSize instead of info.Size() to correctly
			// calculate the size of both files AND directories.
			size, _ = getDirSize(targetPath)

			// Permanently delete from disk
			err := os.RemoveAll(targetPath)
			if err != nil {
				fmt.Printf("%sError: Could not delete '%s': %v%s\n", RED, name, err, RC)
				os.Exit(1)
			}

			found = true
			continue // Skip adding this record to the 'updated' list
		}
		updated = append(updated, r)
	}

	if !found {
		fmt.Printf("%sError: '%s' not found in kepfi trash%s\n", RED, name, RC)
		return
	}

	// Save the updated metadata without the deleted record
	writeRecords(updated)
	fmt.Printf("%s[DONE]%s Permanently removed '%s'. %s cleared\n", GREEN, RC, name, formatSize(size))
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
	cmdStr := fmt.Sprintf("sleep %d && %s -pa -f > /dev/null 2>&1", delay, exe)
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
		fmt.Printf("%sError: Path resolution failed for '%s': %v%s\n", RED, target, err, RC)
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
		destPath = filepath.Join(CACHEDIR, fileName)
		err = os.Rename(absPath, destPath)
		if err != nil {
			fmt.Printf("%sError: Failed to move to temp: %v%s\n", RED, err, RC)
			return
		}
		fmt.Printf("%s[TEMP]%s '%s' moved to /tmp/\n", YELLOW, RC, fileName)
		return
	}

	// Move to standard kepfi trash
	destPath = getUniquePath(trashDir, fileName)
	fileName = filepath.Base(destPath) // metadata
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
	})

	fmt.Printf("%s[TRASH]%s '%s' moved to kepfi trash\n", GREEN, RC, fileName)
}

// purgeEverything permanently deletes all items in the trash and clears metadata.
// It handles partial failures (e.g., root-owned files) without corrupting the database.
func purgeEverything() {
	records := loadRecords()
	size, _ := getDirSize(trashDir)

	// Quick exit if there's nothing to do
	if size == 0 && len(records) == 0 {
		fmt.Printf("kepfi trash is already empty\n")
		return
	}

	// Safety prompt: require confirmation unless --force is used
	if !optForce {
		fmt.Printf("%s%sWARNING: This will permanently delete everything in kepfi trash!%s\n", BOLD, RED, RC)
		fmt.Printf("Confirm action? (y/N): ")
		var response string
		fmt.Scanln(&response)

		if strings.ToLower(response) != "y" {
			fmt.Printf("Action cancelled\n")
			return
		}
	}

	// Track if every single item was successfully removed
	allCleared := true

	// Read top-level items in trash directory
	entries, err := os.ReadDir(trashDir)
	if err != nil {
		fmt.Printf("%sError: Could not read trash folder: %v%s\n", RED, err, RC)
		return
	}

	for _, entry := range entries {
		path := filepath.Join(trashDir, entry.Name())
		err := os.RemoveAll(path)
		if err != nil {
			// Print skip message if a file is locked or owned by root
			fmt.Printf("%sError: Could not delete %s: %v%s\n", RED, entry.Name(), err, RC)
			allCleared = false
		}
	}

	if allCleared {
		// Complete success: wipe metadata and report total cleared size
		_ = os.Remove(metadataFile)
		fmt.Printf("%s[DONE]%s kepfi trash purged. %s%s%s cleared\n", GREEN, RC, YELLOW, formatSize(size), RC)
	} else {
		// Partial failure: filter metadata to only keep records of files still on disk
		fmt.Printf("Some files remain. Updating metadata...\n")

		var remainingRecords []FileRecord
		for _, r := range records {
			if _, err := os.Stat(filepath.Join(trashDir, r.FileName)); err == nil {
				remainingRecords = append(remainingRecords, r)
			}
		}
		writeRecords(remainingRecords)

		fmt.Printf("%sError: Trash not fully cleared%s\n", RED, RC)
	}

	// Ensure the trash directory exists for future use
	_ = os.MkdirAll(trashDir, 0755)
}
