package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	presetsAppStyle = lipgloss.NewStyle().Padding(1, 2)

	presetsTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFDF5")).
				Background(lipgloss.Color("#25A065")).
				Padding(0, 1)
)

// presetItem wraps Preset to implement list.Item interface
type presetItem struct {
	preset *Preset
}

func (i presetItem) Title() string {
	status := " "
	if i.preset.Active {
		status = "✓ "
	}
	return status + i.preset.Name
}

func (i presetItem) Description() string {
	if len(i.preset.Options) == 0 {
		return "No options configured"
	}
	activeCount := 0
	for _, option := range i.preset.Options {
		if option.Enabled {
			activeCount++
		}
	}
	return fmt.Sprintf("%d options (%d active)", len(i.preset.Options), activeCount)
}

func (i presetItem) FilterValue() string {
	return i.preset.Name
}

// NewPresetsView creates a new PresetsView instance
func NewPresetsView() PresetsView {
	// Try to load saved presets from config
	_, _, savedPresets, err := LoadConfig()
	var presets []Preset
	if err != nil || len(savedPresets) == 0 {
		logToFile("Loading default presets (no saved config found)")
		// Use default presets if no saved config
		presets = GetDefaultPresets()
	} else {
		logToFile("Loaded saved presets from config")
		presets = savedPresets
	}

	// Convert presets to list items
	items := make([]list.Item, len(presets))
	for i := range presets {
		items[i] = presetItem{preset: &presets[i]}
	}

	// Create list
	presetsList := list.New(items, list.NewDefaultDelegate(), 0, 0)
	presetsList.SetShowTitle(false) // Hide title
	presetsList.SetShowHelp(false)  // We'll handle help separately

	return PresetsView{
		Presets: presets,
		List:    presetsList,
	}
}

// Update handles input for the PresetsView
func (pv *PresetsView) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := presetsAppStyle.GetFrameSize()
		pv.List.SetSize(msg.Width-h, msg.Height-v)
	case tea.KeyMsg:
		switch msg.String() {
		case " ":
			// Toggle preset active/inactive
			selectedIndex := pv.List.Index()
			if selectedIndex < len(pv.Presets) {
				pv.Presets[selectedIndex].Active = !pv.Presets[selectedIndex].Active
				// Update the list item
				pv.updateListItems()
			}
		case "D":
			// Delete current preset (but not if it's the last one)
			selectedIndex := pv.List.Index()
			if len(pv.Presets) > 1 && selectedIndex < len(pv.Presets) {
				pv.Presets = append(pv.Presets[:selectedIndex], pv.Presets[selectedIndex+1:]...)
				pv.updateListItems()
				// Adjust cursor if needed
				if selectedIndex >= len(pv.Presets) && len(pv.Presets) > 0 {
					pv.List.Select(len(pv.Presets) - 1)
				}
			}
		case "r", "R":
			// Reset to default presets
			pv.Presets = GetDefaultPresets()
			pv.updateListItems()
			pv.List.Select(0)
		default:
			// Let the list handle other keys
			var cmd tea.Cmd
			pv.List, cmd = pv.List.Update(msg)
			return cmd
		}
	}

	return nil
}

// updateListItems synchronizes the list items with the current presets
func (pv *PresetsView) updateListItems() {
	items := make([]list.Item, len(pv.Presets))
	for i := range pv.Presets {
		items[i] = presetItem{preset: &pv.Presets[i]}
	}
	pv.List.SetItems(items)
}

// View renders the PresetsView
func (pv PresetsView) View() string {
	return presetsAppStyle.Render(pv.List.View())
}

// getPresetsHelp returns help text for presets view
func getPresetsHelp() string {
	help := lipgloss.NewStyle().Faint(true)
	return help.Render("N: new preset • Enter: edit • Space: toggle • D: delete • R: reset all • Esc: back • ?: help")
}

// GetActiveOptions returns all enabled options from active presets, handling conflicts
func (pv PresetsView) GetActiveOptions() []Option {
	// Use a map to handle conflicts between presets
	flagMap := make(map[string]Option)

	for _, preset := range pv.Presets {
		if preset.Active {
			for _, option := range preset.Options {
				if option.Enabled {
					flagParts := strings.Fields(option.Flag)
					if len(flagParts) > 0 {
						key := flagParts[0] // e.g., "--format" from "--format=best"
						// Later presets override earlier ones
						flagMap[key] = option
					}
				}
			}
		}
	}

	// Convert map back to slice
	var activeOptions []Option
	for _, option := range flagMap {
		activeOptions = append(activeOptions, option)
	}

	return activeOptions
}

// GetMergedOptions returns options merged with CLI arguments, handling conflicts
func (pv PresetsView) GetMergedOptions(cliArgs []string) []Option {
	// Start with active options from presets
	presetOptions := pv.GetActiveOptions()

	// Create a map to track flags by their key (first word of flag)
	flagMap := make(map[string]Option)

	// Add preset options to map
	for _, option := range presetOptions {
		flagParts := strings.Fields(option.Flag)
		if len(flagParts) > 0 {
			key := flagParts[0] // e.g., "--format" from "--format=best"
			flagMap[key] = option
		}
	}

	// Process CLI arguments and add/override flags
	for i := 0; i < len(cliArgs); i++ {
		arg := cliArgs[i]

		// Skip if it doesn't look like a flag
		if !strings.HasPrefix(arg, "-") {
			continue
		}

		// Handle different flag formats
		var flagKey, fullFlag string
		if strings.Contains(arg, "=") {
			// Format: --flag=value
			parts := strings.SplitN(arg, "=", 2)
			flagKey = parts[0]
			fullFlag = arg
		} else {
			// Format: --flag value (if next arg doesn't start with -)
			flagKey = arg
			if i+1 < len(cliArgs) && !strings.HasPrefix(cliArgs[i+1], "-") {
				fullFlag = arg + " " + cliArgs[i+1]
				i++ // Skip next argument as it's the value
			} else {
				fullFlag = arg
			}
		}

		// Add or override in map
		flagMap[flagKey] = Option{
			Flag:    fullFlag,
			Comment: "From CLI arguments",
			Enabled: true,
		}
	}

	// Convert map back to slice
	var mergedOptions []Option
	for _, option := range flagMap {
		mergedOptions = append(mergedOptions, option)
	}

	// Log merged options for debugging
	if len(cliArgs) > 0 {
		var flagStrings []string
		for _, option := range mergedOptions {
			flagStrings = append(flagStrings, option.Flag)
		}
		logToFile("Merged flags: " + strings.Join(flagStrings, " "))
	}

	return mergedOptions
}

// GetTitle returns the appropriate title for the presets view
func (pv PresetsView) GetTitle() string {
	return "Presets"
}
