package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// renderHeader renders the top header bar
func (m *Model) renderHeader() string {
	title := lipgloss.NewStyle().
		Foreground(ColorPrimary).
		Bold(true).
		Render("⚡ Neon Wallet")

	lockStatus := "🔒 Unlocked"
	if m.locked {
		lockStatus = "🔐 Locked"
	}

	timeStr := time.Now().Format("15:04")
	status := lipgloss.NewStyle().
		Foreground(ColorTextDim).
		Render(fmt.Sprintf("%s │ %s", lockStatus, timeStr))

	headerStyle := lipgloss.NewStyle().
		Width(m.width).
		Padding(0, 1).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(ColorBorder)

	// Space between title and status
	spacer := lipgloss.NewStyle().
		Width(m.width - lipgloss.Width(title) - lipgloss.Width(status) - 4).
		Render("")

	return headerStyle.Render(title + spacer + status)
}

// renderSidebar renders the left navigation sidebar
func (m *Model) renderSidebar() string {
	sidebarWidth := 16

	items := []struct {
		icon  string
		label string
		view  ViewType
	}{
		{"📊", "Dashboard", ViewDashboard},
		{"💼", "Accounts", ViewAccounts},
		{"📤", "Transfer", ViewTransfer},
		{"📜", "History", ViewHistory},
		{"⚙️ ", "Settings", ViewSettings},
	}

	var menuItems []string
	for _, item := range items {
		style := lipgloss.NewStyle().
			Width(sidebarWidth-2).
			Padding(0, 1)

		if item.view == m.currentView {
			style = style.
				Foreground(ColorPrimary).
				Bold(true).
				Background(ColorBgAlt)
		} else {
			style = style.Foreground(ColorText)
		}

		menuItems = append(menuItems, style.Render(fmt.Sprintf("%s %s", item.icon, item.label)))
	}

	sidebarStyle := lipgloss.NewStyle().
		Width(sidebarWidth).
		Height(m.height-4). // Account for header and footer
		BorderStyle(lipgloss.NormalBorder()).
		BorderRight(true).
		BorderForeground(ColorBorder).
		Padding(1, 0)

	return sidebarStyle.Render(lipgloss.JoinVertical(lipgloss.Left, menuItems...))
}

// renderFooter renders the bottom status/help bar
func (m *Model) renderFooter() string {
	help := "↑/k up • ↓/j down • ←/h back • →/l select • q quit • ? help"

	footerStyle := lipgloss.NewStyle().
		Width(m.width).
		Padding(0, 1).
		Foreground(ColorTextDim).
		BorderStyle(lipgloss.NormalBorder()).
		BorderTop(true).
		BorderForeground(ColorBorder)

	return footerStyle.Render(help)
}
