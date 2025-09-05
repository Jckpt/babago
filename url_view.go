package main

import (
	"strings"

	"github.com/76creates/stickers/flexbox"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	// Style definitions for better centering
	styleCentered = lipgloss.NewStyle().Align(lipgloss.Center)
	styleContent  = lipgloss.NewStyle().Align(lipgloss.Center).Padding(1)

	// Button styles
	buttonStyle = lipgloss.NewStyle().
			Bold(true).
			Padding(0, 2).
			Margin(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62"))

	buttonFocusedStyle = buttonStyle.Copy().
				Foreground(lipgloss.Color("205")).
				BorderForeground(lipgloss.Color("205"))

	buttonActiveStyle = buttonStyle.Copy().
				Foreground(lipgloss.Color("15")).
				Background(lipgloss.Color("205")).
				BorderForeground(lipgloss.Color("205"))
)

// NewURLView creates a new URLView instance
func NewURLView() URLView {
	urlInput := textinput.New()
	urlInput.Placeholder = "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
	urlInput.Focus()
	urlInput.CharLimit = 512
	urlInput.Width = 80 // Reasonable width for input

	// Load history from config (presets will be loaded separately)
	urls, names, _, err := LoadConfig()
	if err != nil {
		logToFile("Failed to load config: " + err.Error())
		urls = []string{}
		names = []string{}
	}

	// Create flexbox for centering with background style
	flexBox := flexbox.New(0, 0).SetStyle(styleCentered)

	return URLView{
		URLInput:        urlInput,
		CurrentURL:      "",
		IsValidURL:      false,
		URLHistory:      urls,
		HistoryNames:    names,
		HistoryIndex:    -1,
		IsInHistory:     false,
		FlexBox:         flexBox,
		FocusState:      FocusInput,          // Start with input focused
		LastButtonFocus: FocusDownloadButton, // Default to Download button
	}
}

// Update handles input for the URLView
func (uv *URLView) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Update flexbox size on window resize
		uv.FlexBox.SetWidth(msg.Width)
		uv.FlexBox.SetHeight(msg.Height)
	case tea.KeyMsg:
		switch msg.String() {
		case "up":
			// Handle navigation up
			switch uv.FocusState {
			case FocusDownloadButton, FocusPresetsButton:
				// From buttons go back to input
				uv.FocusState = FocusInput
				uv.URLInput.Focus()
			case FocusInput:
				// Navigate history up when input is focused
				if len(uv.URLHistory) > 0 {
					if !uv.IsInHistory {
						uv.HistoryIndex = len(uv.URLHistory) - 1
						uv.IsInHistory = true
					} else if uv.HistoryIndex > 0 {
						uv.HistoryIndex--
					}
					if uv.HistoryIndex >= 0 && uv.HistoryIndex < len(uv.URLHistory) {
						uv.URLInput.SetValue(uv.URLHistory[uv.HistoryIndex])
						uv.CurrentURL = uv.URLHistory[uv.HistoryIndex]
						uv.IsValidURL = isValidURL(uv.CurrentURL)
					}
				}
			}
		case "down":
			// Handle navigation down
			switch uv.FocusState {
			case FocusInput:
				// Navigate history down when input is focused, or go to first button
				if uv.IsInHistory {
					if uv.HistoryIndex < len(uv.URLHistory)-1 {
						uv.HistoryIndex++
						uv.URLInput.SetValue(uv.URLHistory[uv.HistoryIndex])
						uv.CurrentURL = uv.URLHistory[uv.HistoryIndex]
						uv.IsValidURL = isValidURL(uv.CurrentURL)
					} else {
						// Go beyond history - clear input
						uv.IsInHistory = false
						uv.URLInput.Reset()
						uv.CurrentURL = ""
						uv.IsValidURL = false
					}
				} else {
					// Go to last remembered button
					uv.FocusState = uv.LastButtonFocus
					uv.URLInput.Blur()
				}
			}
		case "left":
			// Handle navigation left
			switch uv.FocusState {
			case FocusPresetsButton:
				uv.FocusState = FocusDownloadButton
				uv.LastButtonFocus = FocusDownloadButton // Remember Download button
			}
		case "right":
			// Handle navigation right
			switch uv.FocusState {
			case FocusDownloadButton:
				uv.FocusState = FocusPresetsButton
				uv.LastButtonFocus = FocusPresetsButton // Remember Presets button
			}
		case "enter", " ":
			// Handle button actions
			switch uv.FocusState {
			case FocusDownloadButton:
				// Return a command to trigger download (will be handled in main.go)
				if uv.CurrentURL != "" && uv.IsValidURL {
					return tea.Cmd(func() tea.Msg {
						return DownloadMsg{
							Progress: DownloadProgress{State: DownloadIdle},
							Done:     false,
						} // Signal download request
					})
				}
			case FocusPresetsButton:
				// Return a command to switch to presets tab
				return tea.Cmd(func() tea.Msg {
					return SwitchTabMsg{Tab: PresetsTab}
				})
			}
		case "ctrl+l":
			// Clear the input
			uv.URLInput.Reset()
			uv.CurrentURL = ""
			uv.IsValidURL = false
			uv.IsInHistory = false

		default:
			// Pass keys to the URL input only if it has focus
			if uv.FocusState == FocusInput {
				uv.URLInput, cmd = uv.URLInput.Update(msg)
				uv.IsInHistory = false // User is typing, exit history mode

				// Auto-save URL on every change
				url := uv.URLInput.Value()
				if url != "" && url != uv.CurrentURL {
					uv.CurrentURL = url
					uv.IsValidURL = isValidURL(url)
				} else if url == "" {
					uv.CurrentURL = ""
					uv.IsValidURL = false
				}
			}
		}
	}

	return cmd
}

// View renders the URLView
func (uv URLView) View() string {
	// Clear existing rows
	uv.FlexBox.SetRows([]*flexbox.Row{})

	// Create input content (no label, no border)
	inputContent := uv.URLInput.View()

	// Status content
	var statusContent string
	if uv.CurrentURL != "" {
		statusStyle := lipgloss.NewStyle()
		if uv.IsValidURL {
			statusStyle = statusStyle.Foreground(lipgloss.Color("10")) // Green

			// Show different text based on whether it's from history
			if uv.IsInHistory && uv.HistoryIndex >= 0 && uv.HistoryIndex < len(uv.HistoryNames) && uv.HistoryNames[uv.HistoryIndex] != "" {
				statusContent = statusStyle.Render(uv.HistoryNames[uv.HistoryIndex])
			} else {
				statusContent = statusStyle.Render("✓ Current URL: " + uv.CurrentURL)
			}
		} else {
			statusStyle = statusStyle.Foreground(lipgloss.Color("9")) // Red
			statusContent = statusStyle.Render("✗ Invalid URL: " + uv.CurrentURL)
		}
	}

	// Create buttons with appropriate styles
	downloadButton := "Download"
	presetsButton := "Presets"

	// Apply styles based on focus
	switch uv.FocusState {
	case FocusDownloadButton:
		downloadButton = buttonFocusedStyle.Render(downloadButton)
		presetsButton = buttonStyle.Render(presetsButton)
	case FocusPresetsButton:
		downloadButton = buttonStyle.Render(downloadButton)
		presetsButton = buttonFocusedStyle.Render(presetsButton)
	default:
		downloadButton = buttonStyle.Render(downloadButton)
		presetsButton = buttonStyle.Render(presetsButton)
	}

	// Create buttons row with space-between layout
	spacer := strings.Repeat(" ", 15) // Space between buttons
	buttonsContent := lipgloss.JoinHorizontal(lipgloss.Left, downloadButton, spacer, presetsButton)

	// Combine input, status and buttons
	fullContent := inputContent
	if statusContent != "" {
		fullContent += "\n\n" + statusContent
	}
	fullContent += "\n\n" + buttonsContent

	// Force recalculate like in the example
	uv.FlexBox.ForceRecalculate()

	// Create 3 rows for vertical centering (top spacer, content, bottom spacer)
	// Top spacer row
	topRow := uv.FlexBox.NewRow().AddCells(
		flexbox.NewCell(1, 1).SetContent(""),
	)

	// Main content row with horizontal centering
	mainRow := uv.FlexBox.NewRow().AddCells(
		flexbox.NewCell(1, 1).SetContent(""), // Left spacer
		flexbox.NewCell(3, 1).
			SetContent(fullContent).
			SetStyle(styleContent), // Main content with padding and centering
		flexbox.NewCell(1, 1).SetContent(""), // Right spacer
	)

	// Bottom spacer row
	bottomRow := uv.FlexBox.NewRow().AddCells(
		flexbox.NewCell(1, 1).SetContent(""),
	)

	// Add all rows to flexbox
	uv.FlexBox.AddRows([]*flexbox.Row{topRow, mainRow, bottomRow})

	return uv.FlexBox.Render()
}

// Focus sets focus to the URLView
func (uv *URLView) Focus() {
	uv.FocusState = FocusInput
	uv.URLInput.Focus()
}

// Blur removes focus from the URLView
func (uv *URLView) Blur() {
	uv.URLInput.Blur()
}

// AddToHistory adds URL to history if it's valid and not already the last entry
func (uv *URLView) AddToHistory(url, name string) {
	if url == "" || !isValidURL(url) {
		return
	}

	// Don't add if it's the same as the last entry
	if len(uv.URLHistory) > 0 && uv.URLHistory[len(uv.URLHistory)-1] == url {
		return
	}

	// Add to history
	uv.URLHistory = append(uv.URLHistory, url)
	uv.HistoryNames = append(uv.HistoryNames, name)

	// Keep max 50 URLs in history
	if len(uv.URLHistory) > 50 {
		uv.URLHistory = uv.URLHistory[1:]
		uv.HistoryNames = uv.HistoryNames[1:]
	}

	// Note: Auto-save will be handled by main.go with complete config
}

// GetURL returns the current URL
func (uv URLView) GetURL() string {
	return uv.CurrentURL
}

// SetURL sets the current URL
func (uv *URLView) SetURL(url string) {
	uv.CurrentURL = url
	uv.URLInput.SetValue(url)
	uv.IsValidURL = isValidURL(url)
}

// isValidURL performs basic URL validation
func isValidURL(url string) bool {
	if len(url) < 7 {
		return false
	}

	// Basic check for http or https protocol
	if url[:7] == "http://" || (len(url) >= 8 && url[:8] == "https://") {
		return true
	}

	return false
}
