package main

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// ytDlpFinishedMsg is sent when yt-dlp process finishes
type ytDlpFinishedMsg struct {
	err               error
	url               string
	downloadStartTime time.Time
}

// ExecuteYtDlpCmd uruchamia yt-dlp używając tea.ExecProcess
func ExecuteYtDlpCmd(url string, options []Option) tea.Cmd {
	downloadStartTime := time.Now()

	// Buduj argumenty komendy
	args := []string{url}

	// Dodaj włączone opcje
	for _, option := range options {
		if option.Enabled {
			flagParts := strings.Fields(option.Flag)
			args = append(args, flagParts...)
		}
	}

	// Loguj komendę
	logToFile("Executing: yt-dlp " + strings.Join(args, " "))

	// Utwórz exec.Cmd
	cmd := exec.Command("yt-dlp", args...)

	// Użyj tea.ExecProcess żeby uruchomić komendę w terminalu
	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		return ytDlpFinishedMsg{
			err:               err,
			url:               url,
			downloadStartTime: downloadStartTime,
		}
	})
}

// findDownloadedFile znajduje najnowszy plik pobrany po określonym czasie
func findDownloadedFile(afterTime time.Time) (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	var newestFile string
	var newestTime time.Time

	err = filepath.WalkDir(currentDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Ignoruj błędy i kontynuuj
		}

		// Sprawdź tylko pliki, nie katalogi
		if d.IsDir() {
			return nil
		}

		// Ignoruj pliki systemowe i konfiguracyjne
		if strings.HasPrefix(d.Name(), ".") || strings.HasSuffix(d.Name(), ".log") {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}

		// Sprawdź czy plik został stworzony po czasie rozpoczęcia pobierania
		if info.ModTime().After(afterTime) {
			if newestFile == "" || info.ModTime().After(newestTime) {
				newestFile = d.Name()
				newestTime = info.ModTime()
			}
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	if newestFile == "" {
		return "", os.ErrNotExist
	}

	return newestFile, nil
}

// runYtDlpDirect executes yt-dlp directly in CLI mode (not through Bubble Tea)
func runYtDlpDirect(url string, options []Option) {
	// Build command arguments
	args := []string{url}

	// Add enabled options
	for _, option := range options {
		if option.Enabled {
			flagParts := strings.Fields(option.Flag)
			args = append(args, flagParts...)
		}
	}

	// Log command
	logToFile("Executing directly: yt-dlp " + strings.Join(args, " "))
	fmt.Println("Executing: yt-dlp " + strings.Join(args, " "))

	// Create and run command directly
	cmd := exec.Command("yt-dlp", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error executing yt-dlp: %v\n", err)
		os.Exit(1)
	}

	logToFile("yt-dlp completed successfully in CLI mode")
}
