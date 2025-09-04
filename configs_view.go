package main

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// NewConfigsView creates a new ConfigsView instance
func NewConfigsView() ConfigsView {
	// Create default presets
	defaultPreset := Preset{
		Name: "Defaults",
		Options: []Option{
			{Flag: "--format=best", Comment: "Download best quality", Enabled: true},
			{Flag: "--write-thumbnail", Comment: "Write thumbnail image", Enabled: true},
			{Flag: "--write-description", Comment: "Write video description", Enabled: true},
		},
		Active: true, // Only Defaults is active by default
	}

	musicPreset := Preset{
		Name: "For Music",
		Options: []Option{
			{Flag: "--extract-audio", Comment: "Extract audio only", Enabled: true},
			{Flag: "--audio-format mp3", Comment: "Convert to MP3", Enabled: true},
			{Flag: "--audio-quality 0", Comment: "Best audio quality", Enabled: true},
			{Flag: "--embed-thumbnail", Comment: "Embed thumbnail in audio", Enabled: true},
		},
		Active: false,
	}

	videoPreset := Preset{
		Name: "For Video",
		Options: []Option{
			{Flag: "--format=best[height<=720]", Comment: "Max 720p video", Enabled: true},
			{Flag: "--write-sub", Comment: "Download subtitles", Enabled: true},
			{Flag: "--sub-lang en", Comment: "English subtitles", Enabled: true},
			{Flag: "--embed-subs", Comment: "Embed subtitles", Enabled: true},
		},
		Active: false,
	}

	// Create input fields
	flagInput := textinput.New()
	flagInput.Placeholder = "--format=best"
	flagInput.CharLimit = 256
	flagInput.Width = 100

	commentInput := textinput.New()
	commentInput.Placeholder = "(optional)"
	commentInput.CharLimit = 256
	commentInput.Width = 100

	return ConfigsView{
		Presets:       []Preset{defaultPreset, musicPreset, videoPreset},
		Cursor:        0,
		ViewMode:      MainView,
		CurrentPreset: -1,
		OptionCursor:  0,
		FlagInput:     flagInput,
		CommentInput:  commentInput,
		InputFocus:    0,
	}
}

// Update handles input for the ConfigsView
func (cv *ConfigsView) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if cv.ViewMode == MainView {
			// Main presets view
			switch msg.String() {
			case "down":
				if cv.Cursor < len(cv.Presets)-1 {
					cv.Cursor++
				}
			case "up":
				if cv.Cursor > 0 {
					cv.Cursor--
				}
			case "enter":
				// Enter preset view
				cv.ViewMode = PresetView
				cv.CurrentPreset = cv.Cursor
				cv.InputFocus = 0
				cv.OptionCursor = 0
				cv.FlagInput.Focus()
				cv.CommentInput.Blur()
			case " ":
				// Toggle preset active/inactive
				cv.Presets[cv.Cursor].Active = !cv.Presets[cv.Cursor].Active
			}
		} else {
			// Preset options view with inputs
			if cv.InputFocus <= 1 {
				// Input fields mode
				switch msg.String() {
				case "down":
					// Switch between inputs
					if cv.InputFocus == 0 {
						cv.InputFocus = 1
						cv.FlagInput.Blur()
						cv.CommentInput.Focus()
					} else {
						// From comment input, go to options list if there are any
						if len(cv.Presets[cv.CurrentPreset].Options) > 0 {
							cv.InputFocus = 2
							cv.CommentInput.Blur()
						} else {
							cv.InputFocus = 0
							cv.CommentInput.Blur()
							cv.FlagInput.Focus()
						}
					}
				case "up":
					// Switch between inputs (reverse)
					if cv.InputFocus == 1 {
						cv.InputFocus = 0
						cv.CommentInput.Blur()
						cv.FlagInput.Focus()
					}
				case "enter":
					// Add new option
					if cv.FlagInput.Value() != "" {
						newOption := Option{
							Flag:    cv.FlagInput.Value(),
							Comment: cv.CommentInput.Value(),
							Enabled: true,
						}
						cv.Presets[cv.CurrentPreset].Options = append(cv.Presets[cv.CurrentPreset].Options, newOption)
						cv.FlagInput.Reset()
						cv.CommentInput.Reset()
						cv.InputFocus = 0
						cv.FlagInput.Focus()
						cv.CommentInput.Blur()
					}
				case "backspace":
					// Go back one level in focus, but only if current input is empty
					if cv.InputFocus == 1 && cv.CommentInput.Value() == "" {
						cv.InputFocus = 0
						cv.CommentInput.Blur()
						cv.FlagInput.Focus()
					} else if cv.InputFocus == 0 && cv.FlagInput.Value() == "" {
						// Return to main view only if flag input is empty
						cv.ViewMode = MainView
						cv.CurrentPreset = -1
						cv.FlagInput.Blur()
						cv.CommentInput.Blur()
						cv.FlagInput.Reset()
						cv.CommentInput.Reset()
					} else {
						// If input has content, let the input handle backspace normally
						if cv.InputFocus == 0 {
							cv.FlagInput, cmd = cv.FlagInput.Update(msg)
						} else {
							cv.CommentInput, cmd = cv.CommentInput.Update(msg)
						}
						return cmd
					}
				default:
					// Pass keys to the focused input
					if cv.InputFocus == 0 {
						cv.FlagInput, cmd = cv.FlagInput.Update(msg)
					} else {
						cv.CommentInput, cmd = cv.CommentInput.Update(msg)
					}
					return cmd
				}
			} else {
				// Options list mode
				switch msg.String() {
				case "down":
					if cv.OptionCursor < len(cv.Presets[cv.CurrentPreset].Options)-1 {
						cv.OptionCursor++
					}
				case "up":
					if cv.OptionCursor > 0 {
						cv.OptionCursor--
					} else {
						// Go back to inputs
						cv.InputFocus = 1
						cv.CommentInput.Focus()
					}
				case "backspace":
					// Return to main view
					cv.ViewMode = MainView
					cv.CurrentPreset = -1
					cv.FlagInput.Blur()
					cv.CommentInput.Blur()
					cv.FlagInput.Reset()
					cv.CommentInput.Reset()
				case " ", "enter":
					// Toggle option enabled/disabled
					preset := &cv.Presets[cv.CurrentPreset]
					if cv.OptionCursor < len(preset.Options) {
						preset.Options[cv.OptionCursor].Enabled = !preset.Options[cv.OptionCursor].Enabled
					}
				case "D":
					// Delete option
					preset := &cv.Presets[cv.CurrentPreset]
					if cv.OptionCursor < len(preset.Options) {
						preset.Options = append(preset.Options[:cv.OptionCursor], preset.Options[cv.OptionCursor+1:]...)
						if cv.OptionCursor >= len(preset.Options) && len(preset.Options) > 0 {
							cv.OptionCursor = len(preset.Options) - 1
						}
						if len(preset.Options) == 0 {
							cv.InputFocus = 0
							cv.FlagInput.Focus()
						}
					}
				}
			}
		}
	}

	return cmd
}

// View renders the ConfigsView
func (cv ConfigsView) View() string {
	if cv.ViewMode == MainView {
		return cv.viewPresetsList()
	} else {
		return cv.viewPresetOptions()
	}
}

// viewPresetsList renders the main presets list
func (cv ConfigsView) viewPresetsList() string {
	s := "Configuration Presets:\n\n\n"

	for i, preset := range cv.Presets {
		cursor := " "
		if cv.Cursor == i {
			cursor = ">"
		}

		status := " "
		if preset.Active {
			status = "✓"
		}

		style := lipgloss.NewStyle()
		if cv.Cursor == i {
			style = style.Bold(true)
		}

		s += fmt.Sprintf("%s [%s] %s\n",
			cursor,
			status,
			style.Render(preset.Name))
	}

	return s
}

// viewPresetOptions renders the options within a preset
func (cv ConfigsView) viewPresetOptions() string {
	preset := cv.Presets[cv.CurrentPreset]

	// Input fields
	flagLabel := "Flag:"
	if cv.InputFocus == 0 {
		flagLabel = lipgloss.NewStyle().Bold(true).Render("→ Flag:")
	}
	s := fmt.Sprintf("%s\n%s\n\n", flagLabel, cv.FlagInput.View())

	commentLabel := "Comment:"
	if cv.InputFocus == 1 {
		commentLabel = lipgloss.NewStyle().Bold(true).Render("→ Comment:")
	}
	s += fmt.Sprintf("%s\n%s\n\n\n", commentLabel, cv.CommentInput.View())

	// Options list
	if len(preset.Options) > 0 {
		title := "Options:"
		if cv.InputFocus == 2 {
			title = lipgloss.NewStyle().Bold(true).Render("→ Options:")
		}
		s += fmt.Sprintf("%s\n", title)

		for i, option := range preset.Options {
			cursor := " "
			if cv.OptionCursor == i && cv.InputFocus == 2 {
				cursor = ">"
			}

			status := " "
			if option.Enabled {
				status = "✓"
			}

			style := lipgloss.NewStyle()
			if cv.OptionCursor == i && cv.InputFocus == 2 {
				style = style.Bold(true)
			}

			s += fmt.Sprintf("%s [%s] %s %s\n",
				cursor,
				status,
				style.Render(option.Flag),
				lipgloss.NewStyle().Faint(true).Render(option.Comment))
		}
		s += "\n"
	}

	return s
}

// GetActiveOptions returns all enabled options from active presets
func (cv ConfigsView) GetActiveOptions() []Option {
	var activeOptions []Option

	for _, preset := range cv.Presets {
		if preset.Active {
			for _, option := range preset.Options {
				if option.Enabled {
					activeOptions = append(activeOptions, option)
				}
			}
		}
	}

	return activeOptions
}

// GetTitle returns the appropriate title for the current view
func (cv ConfigsView) GetTitle() string {
	if cv.ViewMode == PresetView && cv.CurrentPreset >= 0 {
		return fmt.Sprintf("\"%s\" config", cv.Presets[cv.CurrentPreset].Name)
	}
	return "Configs"
}
