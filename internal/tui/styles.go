package tui

import "github.com/charmbracelet/lipgloss"

// Styles contains all the lipgloss styles for the TUI
type Styles struct {
	Title            lipgloss.Style
	SearchPrefix     lipgloss.Style
	SearchInput      lipgloss.Style
	SearchCursor     lipgloss.Style
	ListItem         lipgloss.Style
	ListItemSelected lipgloss.Style
	ListItemMarked   lipgloss.Style
	DatePrefix       lipgloss.Style
	HelpText         lipgloss.Style
	StatusBar        lipgloss.Style
	StatusMessage    lipgloss.Style
	ErrorMessage     lipgloss.Style
	SuccessMessage   lipgloss.Style
	Prompt           lipgloss.Style
	PromptInput      lipgloss.Style
	Muted            lipgloss.Style
	GitBranch        lipgloss.Style
	GitAhead         lipgloss.Style
	GitBehind        lipgloss.Style
	GitDirty         lipgloss.Style
	ColumnHeader     lipgloss.Style
}

// NewStyles creates a new Styles instance with default styling
func NewStyles() Styles {
	return Styles{
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39")). // bright blue
			MarginBottom(1),

		SearchPrefix: lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")). // bright blue
			Bold(true),

		SearchInput: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")), // light gray

		SearchCursor: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Background(lipgloss.Color("240")), // block cursor

		ListItem: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")), // light gray

		ListItemSelected: lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")). // white
			Background(lipgloss.Color("25")). // dark blue
			Bold(true),

		ListItemMarked: lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")). // red
			Bold(true),

		DatePrefix: lipgloss.NewStyle().
			Foreground(lipgloss.Color("242")), // gray

		HelpText: lipgloss.NewStyle().
			Foreground(lipgloss.Color("242")). // gray
			Italic(true),

		StatusBar: lipgloss.NewStyle().
			Foreground(lipgloss.Color("242")).
			MarginTop(1),

		StatusMessage: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")),

		ErrorMessage: lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")). // red
			Bold(true),

		SuccessMessage: lipgloss.NewStyle().
			Foreground(lipgloss.Color("46")). // green
			Bold(true),

		Prompt: lipgloss.NewStyle().
			Foreground(lipgloss.Color("220")). // yellow
			Bold(true).
			MarginTop(1),

		PromptInput: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")),

		Muted: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")), // dark gray

		GitBranch: lipgloss.NewStyle().
			Foreground(lipgloss.Color("81")), // cyan

		GitAhead: lipgloss.NewStyle().
			Foreground(lipgloss.Color("82")), // green

		GitBehind: lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")), // orange

		GitDirty: lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")), // red

		ColumnHeader: lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")). // medium gray
			Bold(true).
			Underline(true),
	}
}
