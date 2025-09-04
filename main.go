package main

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

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
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("Enter", "select/add"),
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
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		TabLeft: key.NewBinding(
			key.WithKeys("["),
			key.WithHelp("[", "previous page"),
		),
		TabRight: key.NewBinding(
			key.WithKeys("]"),
			key.WithHelp("]", "next page"),
		),
	}

	// Create basic model structure
	model := Model{
		Tab:         URLTab, // start with URL tab
		URLView:     NewURLView(),
		ConfigsView: NewConfigsView(),
		Width:       150, // Very wide default
		Height:      40,  // Tall default
		Keys:        keys,
		Help:        help.New(),
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
		return m, nil

	case tea.KeyMsg:
		// Global keys
		switch {
		case key.Matches(msg, m.Keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.Keys.Help):
			m.Help.ShowAll = !m.Help.ShowAll
		case key.Matches(msg, m.Keys.TabLeft):
			// Go to previous tab
			newTab := max(int(m.Tab)-1, 0)
			m.Tab = TabMode(newTab)
			// Reset configs view when switching away from it
			if m.Tab == URLTab {
				m.resetConfigsView()
			}
			m.updateFocus()
			return m, nil
		case key.Matches(msg, m.Keys.TabRight):
			// Go to next tab
			newTab := min(int(m.Tab)+1, 1) // 0=URL, 1=Configs
			m.Tab = TabMode(newTab)
			// Reset configs view when switching away from it
			if m.Tab == URLTab {
				m.resetConfigsView()
			}
			m.updateFocus()
			return m, nil
		}

		// Handle input based on current tab
		switch m.Tab {
		case URLTab:
			// Handle URL view input
			cmd = m.URLView.Update(msg)

		case ConfigsTab:
			// Handle configs view input
			cmd = m.ConfigsView.Update(msg)
		}
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

// resetConfigsView resets the configs view to the main presets list
func (m *Model) resetConfigsView() {
	m.ConfigsView.ViewMode = MainView
	m.ConfigsView.CurrentPreset = -1
	m.ConfigsView.InputFocus = 0
	m.ConfigsView.OptionCursor = 0
	m.ConfigsView.FlagInput.Blur()
	m.ConfigsView.CommentInput.Blur()
	// Reset input values
	m.ConfigsView.FlagInput.Reset()
	m.ConfigsView.CommentInput.Reset()
}

// View renders the entire application
func (m Model) View() string {
	// Log for debugging
	logToFile("Rendering view")

	// Simple tab headers
	urlTabStyle := lipgloss.NewStyle().Bold(true)
	configsTabStyle := lipgloss.NewStyle().Bold(true)

	// Style active tab
	activeStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12")).Background(lipgloss.Color("236"))
	inactiveStyle := lipgloss.NewStyle().Bold(true).Faint(true)

	if m.Tab == URLTab {
		urlTabStyle = activeStyle
	} else {
		urlTabStyle = inactiveStyle
	}

	if m.Tab == ConfigsTab {
		configsTabStyle = activeStyle
	} else {
		configsTabStyle = inactiveStyle
	}

	// Dynamic configs tab title
	configsTitle := m.ConfigsView.GetTitle()

	s := fmt.Sprintf("%s %s\n\n",
		urlTabStyle.Render("[ URL ]"),
		configsTabStyle.Render(fmt.Sprintf("[ %s ]", configsTitle)))

	// Get content for current tab
	var tabContent string
	switch m.Tab {
	case URLTab:
		tabContent = m.URLView.View()
	case ConfigsTab:
		tabContent = m.ConfigsView.View()
	}

	// Simple content rendering
	s += tabContent

	// Add help at the bottom
	s += "\n" + m.Help.View(m.Keys)

	return s
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
	// Initialize model
	model := initialModel()

	p := tea.NewProgram(model, tea.WithAltScreen())

	// Run the program
	_, err := p.Run()
	if err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

func logToFile(msg string) {
	if len(os.Getenv("DEBUG")) > 0 {
		f, err := tea.LogToFile("debug.log", "debug")
		if err != nil {
			fmt.Println("fatal:", err)
			os.Exit(1)
		}
		defer f.Close()
		f.WriteString("TIME: " + time.Now().Format("2006-01-02 15:04:05") + " " + msg + "\n")
	}
}
