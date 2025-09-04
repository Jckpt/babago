package main

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
)

// Option represents a single yt-dlp option
type Option struct {
	Flag    string `json:"flag"`
	Comment string `json:"comment"`
	Enabled bool   `json:"enabled"`
}

// Preset represents a configuration preset
type Preset struct {
	Name    string   `json:"name"`
	Options []Option `json:"options"`
	Active  bool     `json:"active"`
}

// TabMode represents which tab is currently active
type TabMode int

const (
	URLTab TabMode = iota
	ConfigsTab
)

// ViewMode represents the view state
type ViewMode int

const (
	MainView ViewMode = iota
	PresetView
)

// URLView handles the URL input interface
type URLView struct {
	URLInput   textinput.Model
	CurrentURL string
	IsValidURL bool
}

// ConfigsView handles the presets interface
type ConfigsView struct {
	Presets       []Preset
	Cursor        int
	ViewMode      ViewMode
	CurrentPreset int
	OptionCursor  int
	// Input fields for adding new options
	FlagInput    textinput.Model
	CommentInput textinput.Model
	InputFocus   int // 0=flag, 1=comment, 2=options list
}

// AddView handles the adding interface
type AddView struct {
	NameInput     textinput.Model
	CommentInput  textinput.Model
	InputFocus    int      // 0=name, 1=comment
	AddedOptions  []Option // Options added in this session
	OptionsCursor int      // Cursor for options list
	ViewMode      ViewMode // 0=inputs, 1=options list
}

// keyMap defines keybindings for different contexts
type keyMap struct {
	Up        key.Binding
	Down      key.Binding
	Enter     key.Binding
	Space     key.Binding
	Backspace key.Binding
	Delete    key.Binding
	Help      key.Binding
	Quit      key.Binding
	TabLeft   key.Binding
	TabRight  key.Binding
}

// ShortHelp returns keybindings to be shown in the mini help view
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

// FullHelp returns keybindings for the expanded help view
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Enter, k.Space},
		{k.Backspace, k.Delete, k.TabLeft, k.TabRight},
		{k.Help, k.Quit},
	}
}

// Model is the main application model
type Model struct {
	Tab         TabMode
	URLView     URLView
	ConfigsView ConfigsView
	Width       int // Terminal width
	Height      int // Terminal height
	Keys        keyMap
	Help        help.Model
}
