package main

import (
	"github.com/76creates/stickers/flexbox"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
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
	PresetsTab
)

// ViewMode represents the view state
type ViewMode int

const (
	MainView ViewMode = iota
	EditPresetView
	AddOptionViewMode
)

// FocusState represents what element has focus in URLView
type FocusState int

const (
	FocusInput FocusState = iota
	FocusDownloadButton
	FocusPresetsButton
)

// URLView handles the URL input interface
type URLView struct {
	URLInput        textinput.Model
	CurrentURL      string
	IsValidURL      bool
	URLHistory      []string
	HistoryNames    []string // Names for history URLs
	HistoryIndex    int
	IsInHistory     bool
	FlexBox         *flexbox.FlexBox // For centering the input
	FocusState      FocusState       // Which element has focus
	LastButtonFocus FocusState       // Remembers last focused button
}

// PresetsView handles the main presets list interface
type PresetsView struct {
	Presets []Preset
	List    list.Model
}

// PresetView handles editing a single preset
type PresetView struct {
	Preset      *Preset
	OptionsList list.Model // List for options
	// Input fields for adding new options
	FlagInput       textinput.Model
	CommentInput    textinput.Model
	PresetNameInput textinput.Model // For adding new presets
	InputFocus      int             // 0=flag, 1=comment, 2=options list, 3=preset name, 4=add button
}

// keyMap defines keybindings for different contexts
type keyMap struct {
	Up        key.Binding
	Down      key.Binding
	Enter     key.Binding
	Space     key.Binding
	Backspace key.Binding
	Delete    key.Binding
	Download  key.Binding
	Help      key.Binding
	Quit      key.Binding
}

// ShortHelp returns keybindings to be shown in the mini help view
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help}
}

// FullHelp returns keybindings for the expanded help view
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Enter, k.Space},
		{k.Backspace, k.Delete, k.Download},
		{k.Help},
	}
}

// DownloadState represents the current download state
type DownloadState int

const (
	DownloadIdle DownloadState = iota
	DownloadRunning
	DownloadCompleted
	DownloadError
)

// DownloadProgress represents download progress information
type DownloadProgress struct {
	URL        string
	Filename   string
	Percentage string
	Speed      string
	ETA        string
	State      DownloadState
	Output     []string
	Error      string
}

// DownloadMsg is sent when download progress updates
type DownloadMsg struct {
	Progress DownloadProgress
	Line     string
	Done     bool
	Error    error
}

// SwitchTabMsg is sent when switching tabs
type SwitchTabMsg struct {
	Tab TabMode
}

// SwitchToAddOptionMsg is sent when switching to add option view
type SwitchToAddOptionMsg struct{}

// AddOptionMsg is sent when adding an option
type AddOptionMsg struct {
	Flag    string
	Comment string
}

// CancelAddOptionMsg is sent when canceling add option
type CancelAddOptionMsg struct{}

// Model is the main application model
type Model struct {
	Tab           TabMode
	URLView       URLView
	PresetsView   PresetsView
	PresetView    PresetView
	AddOptionView AddOptionView
	CurrentView   ViewMode // MainView for PresetsView, EditPresetView for PresetView
	Download      DownloadProgress
	Width         int // Terminal width
	Height        int // Terminal height
	Keys          keyMap
	Help          help.Model
	ShowHelp      bool // Whether help is expanded
}
