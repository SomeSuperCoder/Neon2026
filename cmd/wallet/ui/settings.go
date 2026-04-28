package ui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/poh-blockchain/cmd/wallet/core"
)

// SettingsView represents the settings view
type SettingsView struct {
	wallet *core.Wallet
}

// NewSettingsView creates a new settings view
func NewSettingsView(wallet *core.Wallet) *SettingsView {
	return &SettingsView{
		wallet: wallet,
	}
}

// Init initializes the settings view
func (v *SettingsView) Init() tea.Cmd {
	return nil
}

// Update handles messages for the settings view
func (v *SettingsView) Update(msg tea.Msg) (ViewModel, tea.Cmd) {
	return v, nil
}

// View renders the settings view
func (v *SettingsView) View() string {
	config := v.wallet.GetConfig()

	var rows []string
	rows = append(rows, TitleStyle.Render("Settings"))
	rows = append(rows, "")

	// RPC Endpoint
	rows = append(rows, lipgloss.NewStyle().Foreground(ColorTextDim).Render("RPC Endpoint:"))
	rows = append(rows, lipgloss.NewStyle().Foreground(ColorText).Render(config.RPCEndpoint))
	rows = append(rows, "")

	// Wallet Path
	rows = append(rows, lipgloss.NewStyle().Foreground(ColorTextDim).Render("Wallet Path:"))
	rows = append(rows, lipgloss.NewStyle().Foreground(ColorText).Render(config.WalletPath))
	rows = append(rows, "")

	// Auto-lock timeout
	rows = append(rows, lipgloss.NewStyle().Foreground(ColorTextDim).Render("Auto-lock Timeout:"))
	timeout := config.AutoLock
	if timeout == 0 {
		timeout = 5 * time.Minute
	}
	rows = append(rows, lipgloss.NewStyle().Foreground(ColorText).Render(fmt.Sprintf("%v", timeout)))
	rows = append(rows, "")

	// Theme
	rows = append(rows, lipgloss.NewStyle().Foreground(ColorTextDim).Render("Theme:"))
	rows = append(rows, lipgloss.NewStyle().Foreground(ColorAccent).Render("Neon (default)"))
	rows = append(rows, "")

	rows = append(rows, SubtitleStyle.Render("Settings are read-only in this version"))

	return PanelStyle.Render(lipgloss.JoinVertical(lipgloss.Left, rows...))
}
