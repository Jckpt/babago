package main

import (
	"strings"

	"github.com/76creates/stickers/flexbox"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	addOptionAppStyle = lipgloss.NewStyle().Padding(1, 2)

	// Button styles matching URL view
	addOptionButtonStyle = lipgloss.NewStyle().
				Bold(true).
				Padding(0, 2).
				Margin(0, 1).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("62"))

	addOptionButtonFocusedStyle = addOptionButtonStyle.Copy().
					Foreground(lipgloss.Color("205")).
					BorderForeground(lipgloss.Color("205"))

	addOptionFocusedLabelStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("170")). // Fuchsia/magenta color used for list selection
					Bold(true)

	// Style for centering
	addOptionStyleCentered = lipgloss.NewStyle().Align(lipgloss.Center)
	addOptionStyleContent  = lipgloss.NewStyle().Align(lipgloss.Center).Padding(3, 6)
)

// AddOptionView handles the add option interface
type AddOptionView struct {
	FlagInput       textinput.Model
	CommentInput    textinput.Model
	InputFocus      int              // 0=flag, 1=comment, 2=add button, 3=cancel button
	LastButtonFocus int              // Remembers last focused button (2 or 3)
	FlexBox         *flexbox.FlexBox // For centering content
}

// NewAddOptionView creates a new AddOptionView instance
func NewAddOptionView() AddOptionView {
	// Create input fields
	flagInput := textinput.New()
	flagInput.Placeholder = "--format=best"
	flagInput.CharLimit = 256
	flagInput.Width = 120
	flagInput.Focus() // Focus on the first input by default

	commentInput := textinput.New()
	commentInput.Placeholder = "(optional)"
	commentInput.CharLimit = 256
	commentInput.Width = 120

	// Create flexbox for centering
	flexBox := flexbox.New(0, 0).SetStyle(addOptionStyleCentered)

	return AddOptionView{
		FlagInput:       flagInput,
		CommentInput:    commentInput,
		InputFocus:      0, // Start with flag input focused
		LastButtonFocus: 2, // Default to Add button
		FlexBox:         flexBox,
	}
}

// Update handles input for the AddOptionView
func (av *AddOptionView) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Update flexbox size on window resize
		av.FlexBox.SetWidth(msg.Width)
		av.FlexBox.SetHeight(msg.Height)
	case tea.KeyMsg:
		switch msg.String() {
		case "down":
			// Navigate down through fields
			if av.InputFocus == 0 {
				// From Flag to Comment
				av.InputFocus = 1
				av.updateInputFocus()
			} else if av.InputFocus == 1 {
				// From Comment to last remembered button
				av.InputFocus = av.LastButtonFocus
				av.updateInputFocus()
			}
			// No action for buttons (2,3) - they are at the bottom
		case "up":
			// Navigate up through fields
			if av.InputFocus == 1 {
				// From Comment to Flag
				av.InputFocus = 0
				av.updateInputFocus()
			} else if av.InputFocus == 2 || av.InputFocus == 3 {
				// From both Add and Cancel buttons to Comment input
				av.InputFocus = 1
				av.updateInputFocus()
			}
		case "left":
			// Navigate between buttons
			if av.InputFocus == 3 { // From Cancel to Add
				av.InputFocus = 2
				av.LastButtonFocus = 2 // Remember Add button
			}
		case "right":
			// Navigate between buttons
			if av.InputFocus == 2 { // From Add to Cancel
				av.InputFocus = 3
				av.LastButtonFocus = 3 // Remember Cancel button
			}
		case "enter":
			// Handle button actions
			if av.InputFocus == 2 { // Add button
				flag := av.FlagInput.Value()
				comment := av.CommentInput.Value()
				if flag != "" {
					return tea.Cmd(func() tea.Msg {
						return AddOptionMsg{Flag: flag, Comment: comment}
					})
				}
			} else if av.InputFocus == 3 { // Cancel button
				return tea.Cmd(func() tea.Msg {
					return CancelAddOptionMsg{}
				})
			}
		default:
			// Pass keys to the focused input
			if av.InputFocus == 0 {
				av.FlagInput, cmd = av.FlagInput.Update(msg)
			} else if av.InputFocus == 1 {
				av.CommentInput, cmd = av.CommentInput.Update(msg)
			}
		}
	}

	return cmd
}

// updateInputFocus sets focus on the correct input field
func (av *AddOptionView) updateInputFocus() {
	switch av.InputFocus {
	case 0:
		av.FlagInput.Focus()
		av.CommentInput.Blur()
	case 1:
		av.FlagInput.Blur()
		av.CommentInput.Focus()
	default:
		av.FlagInput.Blur()
		av.CommentInput.Blur()
	}
}

// Reset clears the input fields and resets focus
func (av *AddOptionView) Reset() {
	av.FlagInput.Reset()
	av.CommentInput.Reset()
	av.InputFocus = 0
	av.updateInputFocus()
	// Don't reset LastButtonFocus - keep memory of last button
}

// View renders the AddOptionView
func (av AddOptionView) View() string {
	// Clear existing rows
	av.FlexBox.SetRows([]*flexbox.Row{})

	// Build content
	content := av.buildContent()

	// Force recalculate
	av.FlexBox.ForceRecalculate()

	// Create 3 rows for proper vertical centering
	// Top spacer row
	topRow := av.FlexBox.NewRow().AddCells(
		flexbox.NewCell(1, 2).SetContent(""),
	)

	// Main content row with horizontal centering
	mainRow := av.FlexBox.NewRow().AddCells(
		flexbox.NewCell(1, 4).SetContent(""), // Left spacer
		flexbox.NewCell(8, 4).
			SetContent(content).
			SetStyle(addOptionStyleContent), // Main content with padding and centering
		flexbox.NewCell(1, 4).SetContent(""), // Right spacer
	)

	// Bottom spacer row
	bottomRow := av.FlexBox.NewRow().AddCells(
		flexbox.NewCell(1, 2).SetContent(""),
	)

	// Add all rows to flexbox
	av.FlexBox.AddRows([]*flexbox.Row{topRow, mainRow, bottomRow})

	return av.FlexBox.Render()
}

// buildContent builds the add option interface content
func (av AddOptionView) buildContent() string {
	var s string

	// Input fields
	flagLabel := "Flag:"
	if av.InputFocus == 0 {
		flagLabel = addOptionFocusedLabelStyle.Render("Flag:")
	}
	s += flagLabel + "\n" + av.FlagInput.View() + "\n\n"

	commentLabel := "Comment:"
	if av.InputFocus == 1 {
		commentLabel = addOptionFocusedLabelStyle.Render("Comment:")
	}
	s += commentLabel + "\n" + av.CommentInput.View() + "\n\n"

	// Buttons row - Add on left, Cancel on right
	addButton := "Add"
	cancelButton := "Cancel"

	if av.InputFocus == 2 {
		addButton = addOptionButtonFocusedStyle.Render("Add")
	} else {
		addButton = addOptionButtonStyle.Render("Add")
	}

	if av.InputFocus == 3 {
		cancelButton = addOptionButtonFocusedStyle.Render("Cancel")
	} else {
		cancelButton = addOptionButtonStyle.Render("Cancel")
	}

	// Create spacer between buttons
	spacer := strings.Repeat(" ", 15) // More space between buttons
	buttonsContent := lipgloss.JoinHorizontal(lipgloss.Left, addButton, spacer, cancelButton)
	s += buttonsContent

	return s
}
