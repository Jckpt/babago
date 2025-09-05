package main

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	presetAppStyle = lipgloss.NewStyle().Padding(1, 2)

	optionsTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFDF5")).
				Background(lipgloss.Color("#25A065")).
				Padding(0, 1)

	focusedLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("170")). // Fuchsia/magenta color used for list selection
				Bold(true)

	// Button styles matching URL view
	addButtonStyle = lipgloss.NewStyle().
			Bold(true).
			Padding(0, 2).
			Margin(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62"))

	addButtonFocusedStyle = addButtonStyle.Copy().
				Foreground(lipgloss.Color("205")).
				BorderForeground(lipgloss.Color("205"))
)

// optionItem wraps Option to implement list.Item interface
type optionItem struct {
	option *Option
}

func (i optionItem) Title() string {
	status := " "
	if i.option.Enabled {
		status = "✓ "
	}
	return status + i.option.Flag
}

func (i optionItem) Description() string {
	if i.option.Comment == "" {
		return "No description"
	}
	return i.option.Comment
}

func (i optionItem) FilterValue() string {
	return i.option.Flag
}

// NewPresetView creates a new PresetView instance
func NewPresetView() PresetView {
	// Create input fields
	flagInput := textinput.New()
	flagInput.Placeholder = "--format=best"
	flagInput.CharLimit = 256
	flagInput.Width = 100

	commentInput := textinput.New()
	commentInput.Placeholder = "(optional)"
	commentInput.CharLimit = 256
	commentInput.Width = 100

	// New preset name input for adding new configs
	presetNameInput := textinput.New()
	presetNameInput.Placeholder = "New preset name"
	presetNameInput.CharLimit = 50
	presetNameInput.Width = 100

	// Create options list
	optionsList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	optionsList.SetShowTitle(false) // Hide title
	optionsList.SetShowHelp(false)  // We'll handle help separately

	return PresetView{
		Preset:          nil,
		OptionsList:     optionsList,
		FlagInput:       flagInput,
		CommentInput:    commentInput,
		PresetNameInput: presetNameInput,
		InputFocus:      0,
	}
}

// SetPreset sets the preset to edit
func (pv *PresetView) SetPreset(preset *Preset) {
	pv.Preset = preset
	// Focus on options list if there are options, otherwise on Add button
	if len(preset.Options) > 0 {
		pv.InputFocus = 2 // Focus on options list
		pv.OptionsList.Select(0)
	} else {
		pv.InputFocus = 4 // Focus on Add button if no options
	}
	pv.FlagInput.Blur()
	pv.CommentInput.Blur()
	pv.PresetNameInput.Blur()
	pv.updateOptionsList()
}

// updateOptionsList synchronizes the options list with the current preset
func (pv *PresetView) updateOptionsList() {
	if pv.Preset == nil {
		pv.OptionsList.SetItems([]list.Item{})
		return
	}

	items := make([]list.Item, len(pv.Preset.Options))
	for i := range pv.Preset.Options {
		items[i] = optionItem{option: &pv.Preset.Options[i]}
	}
	pv.OptionsList.SetItems(items)
}

// SetNewPresetMode puts the view in new preset creation mode
func (pv *PresetView) SetNewPresetMode() {
	pv.Preset = nil
	pv.InputFocus = 5 // Focus on preset name input (new value)
	pv.PresetNameInput.Focus()
	pv.FlagInput.Blur()
	pv.CommentInput.Blur()
}

// Update handles input for the PresetView
func (pv *PresetView) Update(msg tea.Msg) (tea.Cmd, *Preset) {
	var cmd tea.Cmd
	var newPreset *Preset

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Update options list size
		h, v := presetAppStyle.GetFrameSize()
		// Use most of available height like presets_view does
		// Subtract minimal space for Add button at bottom
		availableHeight := msg.Height - v - 4 // Space for Add button at bottom + margin
		if availableHeight < 15 {
			availableHeight = 15 // Higher minimum for better UX
		}
		pv.OptionsList.SetSize(msg.Width-h, availableHeight)
	case tea.KeyMsg:
		if pv.InputFocus == 5 {
			// New preset name input mode
			switch msg.String() {
			case "enter":
				// Create new preset with the entered name
				if pv.PresetNameInput.Value() != "" {
					newPreset = &Preset{
						Name:    pv.PresetNameInput.Value(),
						Active:  false,
						Options: []Option{},
					}
					pv.PresetNameInput.Reset()
					pv.SetPreset(newPreset)
				}
			}
			// Update the input field
			pv.PresetNameInput, cmd = pv.PresetNameInput.Update(msg)
		} else if pv.InputFocus == 2 || pv.InputFocus == 4 {
			// Add button or Options list navigation
			switch msg.String() {
			case " ":
				// Toggle option enabled/disabled (only when in options list)
				if pv.InputFocus == 2 {
					selectedIndex := pv.OptionsList.Index()
					if pv.Preset != nil && selectedIndex < len(pv.Preset.Options) {
						pv.Preset.Options[selectedIndex].Enabled = !pv.Preset.Options[selectedIndex].Enabled
						pv.updateOptionsList()
					}
				}
			case "D":
				// Delete option (only when in options list)
				if pv.InputFocus == 2 {
					selectedIndex := pv.OptionsList.Index()
					if pv.Preset != nil && selectedIndex < len(pv.Preset.Options) {
						pv.Preset.Options = append(pv.Preset.Options[:selectedIndex], pv.Preset.Options[selectedIndex+1:]...)
						pv.updateOptionsList()
						if len(pv.Preset.Options) == 0 {
							pv.InputFocus = 4 // Go to Add button
						} else if selectedIndex >= len(pv.Preset.Options) {
							pv.OptionsList.Select(len(pv.Preset.Options) - 1)
						}
					}
				}
			case "r", "R":
				// Reset current preset to defaults (only when in options list)
				if pv.InputFocus == 2 && pv.Preset != nil {
					defaultPresets := GetDefaultPresets()
					currentName := pv.Preset.Name
					for _, defaultPreset := range defaultPresets {
						if defaultPreset.Name == currentName {
							pv.Preset.Options = make([]Option, len(defaultPreset.Options))
							copy(pv.Preset.Options, defaultPreset.Options)
							pv.updateOptionsList()
							pv.OptionsList.Select(0)
							break
						}
					}
				}
			case "down":
				if pv.InputFocus == 2 && len(pv.Preset.Options) > 0 {
					// In options list, check if at bottom, then go to Add button
					if pv.OptionsList.Index() == len(pv.Preset.Options)-1 {
						pv.InputFocus = 4 // Go to Add button
					} else {
						pv.OptionsList, cmd = pv.OptionsList.Update(msg)
					}
				}
			case "up":
				if pv.InputFocus == 4 {
					// From Add button, go to options list if any
					if len(pv.Preset.Options) > 0 {
						pv.InputFocus = 2
						pv.OptionsList.Select(len(pv.Preset.Options) - 1) // Go to last item
					} else {
						// No options, go to Comment input
						pv.InputFocus = 1
						pv.CommentInput.Focus()
					}
				} else if pv.InputFocus == 2 {
					// In options list, check if at top, then go to inputs
					if pv.OptionsList.Index() == 0 {
						pv.InputFocus = 1 // Go to Comment input
						pv.CommentInput.Focus()
					} else {
						pv.OptionsList, cmd = pv.OptionsList.Update(msg)
					}
				}
			case "enter":
				if pv.InputFocus == 4 {
					// Add button pressed - go to Add Option view
					// This will be handled in main.go
					return tea.Cmd(func() tea.Msg {
						return SwitchToAddOptionMsg{}
					}), newPreset
				}
			default:
				// Let the list handle other keys when focused
				if pv.InputFocus == 2 {
					pv.OptionsList, cmd = pv.OptionsList.Update(msg)
				}
			}
		} else {
			// Input fields mode (InputFocus 0 or 1)
			switch msg.String() {
			case "enter":
				// Add new option
				flag := pv.FlagInput.Value()
				if flag != "" && pv.Preset != nil {
					comment := pv.CommentInput.Value()
					newOption := Option{
						Flag:    flag,
						Comment: comment,
						Enabled: true,
					}
					pv.Preset.Options = append(pv.Preset.Options, newOption)
					pv.updateOptionsList()
					pv.FlagInput.Reset()
					pv.CommentInput.Reset()
					pv.FlagInput.Focus()
					pv.InputFocus = 0
				}
			case "down":
				if pv.InputFocus == 0 {
					pv.InputFocus = 1
					pv.FlagInput.Blur()
					pv.CommentInput.Focus()
				} else if pv.InputFocus == 1 {
					// Go to Add button or options list
					if len(pv.Preset.Options) > 0 {
						pv.InputFocus = 2 // Go to options list
						pv.CommentInput.Blur()
						pv.OptionsList.Select(0)
					} else {
						pv.InputFocus = 4 // Go to Add button
						pv.CommentInput.Blur()
					}
				}
			case "up":
				if pv.InputFocus == 1 {
					pv.InputFocus = 0
					pv.CommentInput.Blur()
					pv.FlagInput.Focus()
				} else if pv.InputFocus == 0 {
					// From Flag input, go to bottom of the screen (circular navigation)
					if len(pv.Preset.Options) > 0 {
						// Go to last option in list
						pv.InputFocus = 2
						pv.FlagInput.Blur()
						pv.OptionsList.Select(len(pv.Preset.Options) - 1)
					} else {
						// No options, go to Add button
						pv.InputFocus = 4
						pv.FlagInput.Blur()
					}
				}
			default:
				// Handle input field updates
				if pv.InputFocus == 0 {
					pv.FlagInput, cmd = pv.FlagInput.Update(msg)
				} else if pv.InputFocus == 1 {
					pv.CommentInput, cmd = pv.CommentInput.Update(msg)
				}
			}
		}
	}

	return cmd, newPreset
}

// View renders the PresetView
func (pv PresetView) View() string {
	// Handle new preset creation mode
	if pv.InputFocus == 5 { // New value for preset name input
		return presetAppStyle.Render(pv.viewNewPresetInput())
	}

	if pv.Preset == nil {
		return presetAppStyle.Render("No preset selected")
	}

	// Build content
	content := pv.buildPresetContent()

	return presetAppStyle.Render(content)
}

// getPresetHelp returns context-aware help text for preset view
func getPresetHelp(inputFocus int) string {
	help := lipgloss.NewStyle().Faint(true)

	switch inputFocus {
	case 0, 1: // Input fields
		return help.Render("Enter: add option • Tab/↓: next field • Esc: back • ?: help • q: quit")
	case 2: // Options list
		return help.Render("Space: toggle • D: delete • R: reset • ↑/↓: navigate • Esc: back • ?: help • q: quit")
	case 3: // New preset name
		return help.Render("Enter: create • Esc: cancel • ?: help • q: quit")
	default:
		return help.Render("?: help • q: quit")
	}
}

// buildPresetContent builds the main preset editing content
func (pv PresetView) buildPresetContent() string {
	var s string

	// Options list first
	if len(pv.Preset.Options) > 0 {
		s += pv.OptionsList.View()
	} else {
		s += "No options yet. Click Add to create one!"
	}

	// Add button below the list
	s += "\n\n"
	addButton := "Add Option"
	if pv.InputFocus == 4 {
		addButton = addButtonFocusedStyle.Render("Add Option")
	} else {
		addButton = addButtonStyle.Render("Add Option")
	}
	s += addButton

	return s
}

// viewNewPresetInput renders the new preset creation interface
func (pv PresetView) viewNewPresetInput() string {
	nameLabel := "Preset Name:"
	if pv.InputFocus == 3 {
		nameLabel = focusedLabelStyle.Render("Preset Name:")
	}

	content := fmt.Sprintf("Create New Preset:\n\n%s\n%s\n\n", nameLabel, pv.PresetNameInput.View())
	content += "Press Enter to create preset\n"
	content += "Press Esc to cancel\n"

	return content
}

// GetTitle returns the appropriate title for the current view
func (pv PresetView) GetTitle() string {
	if pv.InputFocus == 3 {
		return "New Preset"
	} else if pv.Preset != nil {
		return fmt.Sprintf("\"%s\" preset", pv.Preset.Name)
	}
	return "Preset"
}
