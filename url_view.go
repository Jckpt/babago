package main

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// NewURLView creates a new URLView instance
func NewURLView() URLView {
	urlInput := textinput.New()
	urlInput.Placeholder = "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
	urlInput.Focus()
	urlInput.CharLimit = 512
	urlInput.Width = 120 // Much wider

	return URLView{
		URLInput:   urlInput,
		CurrentURL: "",
		IsValidURL: false,
	}
}

// Update handles input for the URLView
func (uv *URLView) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			// Save the URL
			url := uv.URLInput.Value()
			if url != "" {
				uv.CurrentURL = url
				uv.IsValidURL = isValidURL(url)
			}

		case "ctrl+l":
			// Clear the input
			uv.URLInput.Reset()
			uv.CurrentURL = ""
			uv.IsValidURL = false

		default:
			// Pass all other keys to the URL input
			uv.URLInput, cmd = uv.URLInput.Update(msg)
		}
	}

	return cmd
}

// View renders the URLView
func (uv URLView) View() string {
	// URL input
	inputLabel := "URL:"
	inputLabel = lipgloss.NewStyle().Bold(true).Render("→ " + inputLabel)
	s := fmt.Sprintf("%s\n%s\n\n\n", inputLabel, uv.URLInput.View())

	// Current URL status
	if uv.CurrentURL != "" {
		statusStyle := lipgloss.NewStyle()
		if uv.IsValidURL {
			statusStyle = statusStyle.Foreground(lipgloss.Color("10")) // Green
			s += statusStyle.Render("✓ Current URL: " + uv.CurrentURL + "\n\n")
		} else {
			statusStyle = statusStyle.Foreground(lipgloss.Color("9")) // Red
			s += statusStyle.Render("✗ Invalid URL: " + uv.CurrentURL + "\n\n")
		}
	}

	return s
}

// Focus sets focus to the URLView
func (uv *URLView) Focus() {
	uv.URLInput.Focus()
}

// Blur removes focus from the URLView
func (uv *URLView) Blur() {
	uv.URLInput.Blur()
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
