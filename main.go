package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Global variable to store CLI arguments
var cliArgs []string

// initialModel creates the initial application model
func initialModel() Model {
	// Create keybindings
	keys := keyMap{
		Up: key.NewBinding(
			key.WithKeys("up"),
			key.WithHelp("↑", "navigation"),
		),
		Down: key.NewBinding(
			key.WithKeys("down"),
			key.WithHelp("↓", "navigation"),
		),
		Space: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("Space", "toggle"),
		),
		Backspace: key.NewBinding(
			key.WithKeys("backspace"),
			key.WithHelp("Backspace", "go back"),
		),
		Delete: key.NewBinding(
			key.WithKeys("D"),
			key.WithHelp("D", "delete"),
		),
		Download: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("Enter", "download"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("", ""),
		),
		// Removed tab switching keys
	}

	// Create basic model structure
	model := Model{
		Tab:           URLTab, // start with URL tab
		URLView:       NewURLView(),
		PresetsView:   NewPresetsView(),
		PresetView:    NewPresetView(),
		AddOptionView: NewAddOptionView(),
		CurrentView:   MainView,
		Width:         150, // Very wide default
		Height:        40,  // Tall default
		Keys:          keys,
		Help:          help.New(),
		ShowHelp:      false, // Help starts collapsed
	}

	// Set initial focus
	model.updateFocus()

	return model
}

// Init initializes the application
func (m Model) Init() tea.Cmd {
	return tea.EnterAltScreen
}

// Update handles all input and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {

	// Handle window size changes
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		// Update help width
		m.Help.Width = msg.Width
		// Update URLView flexbox size
		m.URLView.Update(msg)
		// Update PresetsView list size
		m.PresetsView.Update(msg)
		// Update PresetView list size
		m.PresetView.Update(msg)
		// Update AddOptionView flexbox size
		m.AddOptionView.Update(msg)
		return m, nil

	case tea.KeyMsg:
		// Global keys
		switch {
		case key.Matches(msg, m.Keys.Help):
			m.Help.ShowAll = !m.Help.ShowAll
			m.ShowHelp = m.Help.ShowAll
		case key.Matches(msg, m.Keys.Download):
			// Start download only on URL tab if URL is provided
			if m.Tab == URLTab && m.URLView.CurrentURL != "" {
				// Get merged options (presets + CLI args)
				mergedOptions := m.PresetsView.GetMergedOptions(cliArgs)
				return m, ExecuteYtDlpCmd(m.URLView.CurrentURL, mergedOptions)
			}
			// Don't handle Enter for other tabs - let them handle it themselves
			if m.Tab != URLTab {
				break // Exit this switch and let tab-specific handlers deal with Enter
			}
		}

		// Handle input based on current tab
		switch m.Tab {
		case URLTab:
			// Handle URL view input
			if msg.String() == "esc" {
				return m, tea.Quit
			}
			cmd = m.URLView.Update(msg)

		case PresetsTab:
			// Handle presets view input
			if m.CurrentView == MainView {
				// Handle main presets list
				switch msg.String() {
				case "esc":
					// Go back to URL tab
					m.Tab = URLTab
					m.resetPresetsView()
					m.updateFocus()
					return m, nil
				case "enter":
					// Enter preset editing mode
					m.CurrentView = EditPresetView
					selectedIndex := m.PresetsView.List.Index()
					if selectedIndex < len(m.PresetsView.Presets) {
						selectedPreset := &m.PresetsView.Presets[selectedIndex]
						m.PresetView.SetPreset(selectedPreset)
					}
				case "n", "N":
					// Create new preset
					m.CurrentView = EditPresetView
					m.PresetView.SetNewPresetMode()
				default:
					_ = m.PresetsView.Update(msg)
				}
			} else if m.CurrentView == EditPresetView {
				// Handle preset editing
				switch msg.String() {
				case "esc":
					// Go back to main view
					m.CurrentView = MainView
				default:
					updateCmd, newPreset := m.PresetView.Update(msg)
					if newPreset != nil {
						m.PresetsView.Presets = append(m.PresetsView.Presets, *newPreset)
					}
					if updateCmd != nil {
						cmd = updateCmd
					}
				}
			} else if m.CurrentView == AddOptionViewMode {
				// Handle add option view
				switch msg.String() {
				case "esc":
					// Go back to edit view
					return m, tea.Cmd(func() tea.Msg {
						return CancelAddOptionMsg{}
					})
				default:
					// Pass to AddOptionView
					cmd = m.AddOptionView.Update(msg)
				}
			}
			// Auto-save config after any changes
			AutoSaveConfig(&m.URLView, &m.PresetsView)
		}

	// Handle yt-dlp process finished
	case ytDlpFinishedMsg:
		if msg.err != nil {
			logToFile("yt-dlp finished with error: " + msg.err.Error())
		} else {
			logToFile("yt-dlp finished successfully")

			// Try to find the downloaded file
			filename, err := findDownloadedFile(msg.downloadStartTime)
			if err != nil {
				logToFile("Could not find downloaded file: " + err.Error())
				// Use fallback name
				filename = generateVideoName(msg.url)
			} else {
				logToFile("Found downloaded file: " + filename)
			}

			// Add to history with actual filename
			m.URLView.AddToHistory(msg.url, filename)

			// Auto-save complete config
			AutoSaveConfig(&m.URLView, &m.PresetsView)
		}
		// Continue running the app after yt-dlp finishes
		return m, nil

	// Handle download messages (keeping for compatibility)
	case DownloadMsg:
		// If this is a download request from button (not actual progress)
		if !msg.Done && msg.Progress.State == DownloadIdle {
			// Start download if URL is valid
			if m.URLView.CurrentURL != "" {
				mergedOptions := m.PresetsView.GetMergedOptions(cliArgs)
				return m, ExecuteYtDlpCmd(m.URLView.CurrentURL, mergedOptions)
			}
		}

		m.Download = msg.Progress
		if msg.Done {
			if msg.Error != nil {
				m.Download.State = DownloadError
				m.Download.Error = msg.Error.Error()
			} else if m.Download.State != DownloadError {
				m.Download.State = DownloadCompleted
			}
		}
		return m, nil

	// Handle tab switching from buttons
	case SwitchTabMsg:
		m.Tab = msg.Tab
		// Reset presets view when switching away from it
		if m.Tab == URLTab {
			m.resetPresetsView()
		}
		m.updateFocus()
		return m, nil

	// Handle switching to add option view
	case SwitchToAddOptionMsg:
		if m.Tab == PresetsTab && m.CurrentView == EditPresetView {
			m.CurrentView = AddOptionViewMode
			m.AddOptionView.Reset() // Reset and focus on flag input
		}
		return m, nil

	// Handle adding option
	case AddOptionMsg:
		if m.Tab == PresetsTab && m.CurrentView == AddOptionViewMode && m.PresetView.Preset != nil {
			newOption := Option{
				Flag:    msg.Flag,
				Comment: msg.Comment,
				Enabled: true,
			}
			m.PresetView.Preset.Options = append(m.PresetView.Preset.Options, newOption)
			m.PresetView.updateOptionsList()
			// Go back to Edit view
			m.CurrentView = EditPresetView
			m.PresetView.InputFocus = 4 // Focus on Add button
			AutoSaveConfig(&m.URLView, &m.PresetsView)
		}
		return m, nil

	// Handle canceling add option
	case CancelAddOptionMsg:
		if m.Tab == PresetsTab && m.CurrentView == AddOptionViewMode {
			// Go back to Edit view
			m.CurrentView = EditPresetView
			// Focus on options list if there are options, otherwise on Add button
			if m.PresetView.Preset != nil && len(m.PresetView.Preset.Options) > 0 {
				m.PresetView.InputFocus = 2 // Focus on options list
				m.PresetView.OptionsList.Select(0)
			} else {
				m.PresetView.InputFocus = 4 // Focus on Add button
			}
		}
		return m, nil
	}

	return m, cmd
}

// updateFocus sets focus based on current tab
func (m *Model) updateFocus() {
	// Blur all first
	m.URLView.Blur()

	// Focus current tab
	switch m.Tab {
	case URLTab:
		m.URLView.Focus()
		// ConfigsTab doesn't need special focus
	}
}

// resetPresetsView resets the presets view to the main presets list
func (m *Model) resetPresetsView() {
	m.CurrentView = MainView
	m.PresetView.InputFocus = 0
	m.PresetView.FlagInput.Blur()
	m.PresetView.CommentInput.Blur()
	m.PresetView.PresetNameInput.Blur()
	// Reset input values
	m.PresetView.FlagInput.Reset()
	m.PresetView.CommentInput.Reset()
	m.PresetView.PresetNameInput.Reset()
	// Reset options list selection
	m.PresetView.OptionsList.Select(0)
}

// renderDownloadProgress renders the download progress information
func (m Model) renderDownloadProgress() string {
	switch m.Download.State {
	case DownloadRunning:
		style := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
		progress := "Downloading..." + "\n"
		if m.Download.Percentage != "" {
			progress = fmt.Sprintf("Downloading: %s", m.Download.Percentage)
		}
		if m.Download.Speed != "" {
			progress += fmt.Sprintf(" at %s", m.Download.Speed)
		}
		if m.Download.ETA != "" {
			progress += fmt.Sprintf(" (ETA: %s)", m.Download.ETA)
		}
		if m.Download.Filename != "" {
			progress += fmt.Sprintf("\nFile: %s", m.Download.Filename)
		}
		return style.Render(progress)

	case DownloadCompleted:
		style := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10"))
		result := "✅ Download completed!"
		if m.Download.Filename != "" {
			result += fmt.Sprintf("\nFile: %s", m.Download.Filename)
		}
		return style.Render(result)

	case DownloadError:
		style := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("9"))
		return style.Render(fmt.Sprintf("❌ Download error: %s", m.Download.Error))

	default:
		return ""
	}
}

// View renders the entire application
func (m Model) View() string {
	// Log for debugging
	logToFile("Rendering view")
	logToFile("Download command: ")

	// Remove header with tabs - no tab display needed
	var s string

	// Get content for current tab
	var tabContent string
	switch m.Tab {
	case URLTab:
		tabContent = m.URLView.View()
	case PresetsTab:
		if m.CurrentView == MainView {
			tabContent = m.PresetsView.View()
		} else if m.CurrentView == EditPresetView {
			tabContent = m.PresetView.View()
		} else if m.CurrentView == AddOptionViewMode {
			tabContent = m.AddOptionView.View()
		}
	}

	// Simple content rendering
	s = tabContent

	// Add help first
	if m.Tab == URLTab {
		// Show URL help always with Esc: quit
		s += "\n" + getURLHelpText(m.ShowHelp)
	} else if m.Tab == PresetsTab {
		if m.CurrentView == MainView {
			s += "\n" + getPresetsHelpText(m.ShowHelp)
		} else if m.CurrentView == AddOptionViewMode {
			s += "\n" + getAddOptionHelpText(m.AddOptionView.InputFocus, m.ShowHelp)
		} else {
			s += "\n" + getPresetHelpText(m.PresetView.InputFocus, m.ShowHelp)
		}
	}

	// Add download progress at the very bottom if downloading
	if m.Download.State != DownloadIdle {
		s += "\n" + m.renderDownloadProgress()
	}

	return s
}

// getPresetsHelpText returns help text for presets view
func getPresetsHelpText(showHelp bool) string {
	help := lipgloss.NewStyle().Faint(true)
	if showHelp {
		return help.Render("N: new preset • Enter: edit • Space: toggle • D: delete • R: reset all • Esc: back • ?: hide help")
	}
	return help.Render("?: help")
}

// getPresetHelpText returns context-aware help text for preset view
func getPresetHelpText(inputFocus int, showHelp bool) string {
	help := lipgloss.NewStyle().Faint(true)

	if !showHelp {
		return help.Render("?: help")
	}

	switch inputFocus {
	case 0, 1: // Input fields
		return help.Render("Enter: add option • Tab/↓: next field • Esc: back • ?: hide help")
	case 2: // Options list
		return help.Render("Space: toggle • D: delete • R: reset • ↑/↓: navigate • Esc: back • ?: hide help")
	case 3: // New preset name
		return help.Render("Enter: create • Esc: cancel • ?: hide help")
	default:
		return help.Render("?: hide help")
	}
}

// getAddOptionHelpText returns help text for add option view
func getAddOptionHelpText(inputFocus int, showHelp bool) string {
	help := lipgloss.NewStyle().Faint(true)

	if !showHelp {
		return help.Render("?: help")
	}

	switch inputFocus {
	case 0, 1: // Input fields
		return help.Render("Enter: add option • ↑/↓: navigate • Esc: cancel • ?: hide help")
	case 2: // Add button
		return help.Render("Enter: add option • ↑/↓: navigate • Esc: cancel • ?: hide help")
	case 3: // Cancel button
		return help.Render("Enter: cancel • ↑/↓: navigate • Esc: cancel • ?: hide help")
	default:
		return help.Render("?: hide help")
	}
}

func getURLHelpText(showHelp bool) string {
	help := lipgloss.NewStyle().Faint(true)
	// Always show Esc: quit and ? toggle text; when expanded, add details
	if !showHelp {
		return help.Render("Esc: quit • ?: help")
	}
	return help.Render("Esc: quit • Enter: download • →/←: switch button • ?: hide help")
}

// Simple styles - no complex borders needed
var (
	docStyle = lipgloss.NewStyle()
)

// Helper functions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func main() {
	// Parse CLI arguments (skip program name)
	cliArgs = os.Args[1:]

	// If CLI arguments are provided, run yt-dlp directly without TUI
	if len(cliArgs) > 0 {
		runDirectYtDlp(cliArgs)
		return
	}

	// Initialize model for TUI mode
	model := initialModel()

	p := tea.NewProgram(model, tea.WithAltScreen())

	// Run the program
	_, err := p.Run()
	if err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

// runDirectYtDlp executes yt-dlp directly with CLI arguments, merged with saved config
func runDirectYtDlp(args []string) {
	// Log CLI execution mode
	logToFile("Running in CLI mode with args: " + strings.Join(args, " "))

	// Find URL in arguments (first argument that looks like a URL)
	var url string
	var nonUrlArgs []string

	for _, arg := range args {
		if isValidURL(arg) {
			url = arg
		} else {
			nonUrlArgs = append(nonUrlArgs, arg)
		}
	}

	if url == "" {
		fmt.Println("Error: No valid URL found in arguments")
		fmt.Println("Usage: babago [URL] [yt-dlp options...]")
		os.Exit(1)
	}

	// Load saved configuration
	presetsView := NewPresetsView()

	// Get merged options (saved config + CLI args)
	mergedOptions := presetsView.GetMergedOptions(nonUrlArgs)

	// Execute yt-dlp directly
	runYtDlpDirect(url, mergedOptions)
}

// generateVideoName generates a simple name for video based on URL
func generateVideoName(url string) string {
	// Simple name generation based on URL
	if strings.Contains(url, "youtube.com") || strings.Contains(url, "youtu.be") {
		// Extract video ID from YouTube URL
		if strings.Contains(url, "v=") {
			parts := strings.Split(url, "v=")
			if len(parts) > 1 {
				videoID := strings.Split(parts[1], "&")[0]
				if len(videoID) > 8 {
					return "YouTube_" + videoID[:8]
				}
			}
		} else if strings.Contains(url, "youtu.be/") {
			parts := strings.Split(url, "youtu.be/")
			if len(parts) > 1 {
				videoID := strings.Split(parts[1], "?")[0]
				if len(videoID) > 8 {
					return "YouTube_" + videoID[:8]
				}
			}
		}
		return "YouTube_video"
	}

	// For other URLs, use domain
	if strings.Contains(url, "://") {
		parts := strings.Split(url, "://")
		if len(parts) > 1 {
			domain := strings.Split(parts[1], "/")[0]
			domain = strings.Replace(domain, "www.", "", 1)
			return strings.Title(domain) + "_video"
		}
	}

	return "video"
}

func logToFile(msg string) {
	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		fmt.Println("fatal:", err)
		os.Exit(1)
	}
	defer f.Close()
	f.WriteString("TIME: " + time.Now().Format("2006-01-02 15:04:05") + " " + msg + "\n")
}
