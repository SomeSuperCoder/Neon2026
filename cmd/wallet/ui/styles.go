package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// Color palette - neon-inspired modern theme
var (
	// Primary colors
	ColorPrimary   = lipgloss.Color("#00D9FF") // Neon cyan
	ColorSecondary = lipgloss.Color("#FF00FF") // Neon magenta
	ColorAccent    = lipgloss.Color("#00FF9F") // Neon green

	// Status colors
	ColorSuccess = lipgloss.Color("#00FF9F") // Green
	ColorError   = lipgloss.Color("#FF0055") // Red
	ColorWarning = lipgloss.Color("#FFD700") // Gold
	ColorInfo    = lipgloss.Color("#00D9FF") // Cyan

	// UI colors
	ColorBorder  = lipgloss.Color("#3A3A3A") // Dark gray
	ColorText    = lipgloss.Color("#E0E0E0") // Light gray
	ColorTextDim = lipgloss.Color("#808080") // Medium gray
	ColorBg      = lipgloss.Color("#1A1A1A") // Very dark gray
	ColorBgAlt   = lipgloss.Color("#2A2A2A") // Dark gray
)

// Common styles
var (
	// Panel style for content areas
	PanelStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Padding(1, 2)

	// Title style for section headers
	TitleStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true).
			MarginBottom(1)

	// Subtitle style
	SubtitleStyle = lipgloss.NewStyle().
			Foreground(ColorTextDim).
			Italic(true)

	// Success message style
	SuccessStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Bold(true)

	// Error message style
	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorError).
			Bold(true)

	// Warning message style
	WarningStyle = lipgloss.NewStyle().
			Foreground(ColorWarning).
			Bold(true)

	// Info message style
	InfoStyle = lipgloss.NewStyle().
			Foreground(ColorInfo)

	// Button style
	ButtonStyle = lipgloss.NewStyle().
			Foreground(ColorText).
			Background(ColorBgAlt).
			Padding(0, 2).
			MarginRight(1).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder)

	// Active button style
	ButtonActiveStyle = lipgloss.NewStyle().
				Foreground(ColorBg).
				Background(ColorPrimary).
				Padding(0, 2).
				MarginRight(1).
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(ColorPrimary).
				Bold(true)

	// Input field style
	InputStyle = lipgloss.NewStyle().
			Foreground(ColorText).
			Background(ColorBgAlt).
			Padding(0, 1).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(ColorBorder)

	// Focused input field style
	InputFocusedStyle = lipgloss.NewStyle().
				Foreground(ColorText).
				Background(ColorBgAlt).
				Padding(0, 1).
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(ColorPrimary)

	// Table header style
	TableHeaderStyle = lipgloss.NewStyle().
				Foreground(ColorPrimary).
				Bold(true).
				BorderStyle(lipgloss.NormalBorder()).
				BorderBottom(true).
				BorderForeground(ColorBorder).
				Padding(0, 1)

	// Table cell style
	TableCellStyle = lipgloss.NewStyle().
			Foreground(ColorText).
			Padding(0, 1)

	// Table cell alternate row style
	TableCellAltStyle = lipgloss.NewStyle().
				Foreground(ColorText).
				Background(ColorBgAlt).
				Padding(0, 1)

	// Selected table row style
	TableRowSelectedStyle = lipgloss.NewStyle().
				Foreground(ColorBg).
				Background(ColorPrimary).
				Bold(true).
				Padding(0, 1)
)

// Helper functions for styled text

// FormatAddress truncates and styles an address
func FormatAddress(addr string) string {
	if len(addr) <= 12 {
		return lipgloss.NewStyle().Foreground(ColorAccent).Render(addr)
	}
	truncated := addr[:6] + "..." + addr[len(addr)-4:]
	return lipgloss.NewStyle().Foreground(ColorAccent).Render(truncated)
}

// FormatBalance formats and styles a balance
func FormatBalance(balance int64) string {
	return lipgloss.NewStyle().
		Foreground(ColorSuccess).
		Bold(true).
		Render(fmt.Sprintf("%d", balance))
}

// FormatAmount formats and styles an amount with direction
func FormatAmount(amount int64, incoming bool) string {
	symbol := "↑"
	color := ColorError
	if incoming {
		symbol = "↓"
		color = ColorSuccess
	}

	return lipgloss.NewStyle().
		Foreground(color).
		Render(fmt.Sprintf("%s %d", symbol, amount))
}

// FormatTimestamp formats and styles a timestamp
func FormatTimestamp(t time.Time) string {
	return lipgloss.NewStyle().
		Foreground(ColorTextDim).
		Render(t.Format("2006-01-02 15:04:05"))
}

// FormatSignature truncates and styles a transaction signature
func FormatSignature(sig string) string {
	if len(sig) <= 16 {
		return lipgloss.NewStyle().Foreground(ColorInfo).Render(sig)
	}
	truncated := sig[:8] + "..." + sig[len(sig)-8:]
	return lipgloss.NewStyle().Foreground(ColorInfo).Render(truncated)
}
