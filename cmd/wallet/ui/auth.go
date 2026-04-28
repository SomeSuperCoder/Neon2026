package ui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/poh-blockchain/cmd/wallet/core"
)

// AuthView represents the authentication/password entry view
type AuthView struct {
	wallet         *core.Wallet
	password       string
	failedAttempts int
	lockUntil      time.Time
	err            error
	mode           AuthMode
}

// AuthMode represents the authentication mode
type AuthMode int

const (
	AuthModeUnlock AuthMode = iota
	AuthModeCreate
	AuthModeRestore
)

// NewAuthView creates a new authentication view
func NewAuthView(wallet *core.Wallet, mode AuthMode) *AuthView {
	return &AuthView{
		wallet:         wallet,
		mode:           mode,
		failedAttempts: 0,
	}
}

// Init initializes the auth view
func (v *AuthView) Init() tea.Cmd {
	return nil
}

// Update handles messages for the auth view
func (v *AuthView) Update(msg tea.Msg) (ViewModel, tea.Cmd) {
	// Check if locked
	if time.Now().Before(v.lockUntil) {
		return v, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			return v, v.authenticate()
		case "backspace":
			if len(v.password) > 0 {
				v.password = v.password[:len(v.password)-1]
			}
		case "ctrl+c", "esc":
			return v, tea.Quit
		default:
			// Add character to password
			if len(msg.String()) == 1 {
				v.password += msg.String()
			}
		}

	case authResultMsg:
		if msg.success {
			// Authentication successful - this would transition to main app
			return v, nil
		} else {
			v.failedAttempts++
			v.err = fmt.Errorf(msg.error)

			// Lock after 3 failed attempts
			if v.failedAttempts >= 3 {
				v.lockUntil = time.Now().Add(30 * time.Second)
				v.err = fmt.Errorf("too many failed attempts, locked for 30 seconds")
			}

			v.password = ""
		}
	}

	return v, nil
}

// View renders the auth view
func (v *AuthView) View() string {
	// Check if locked
	if time.Now().Before(v.lockUntil) {
		remaining := time.Until(v.lockUntil).Seconds()
		return v.renderLocked(int(remaining))
	}

	switch v.mode {
	case AuthModeUnlock:
		return v.renderUnlock()
	case AuthModeCreate:
		return v.renderCreate()
	case AuthModeRestore:
		return v.renderRestore()
	default:
		return v.renderUnlock()
	}
}

// renderUnlock renders the unlock screen
func (v *AuthView) renderUnlock() string {
	var rows []string
	rows = append(rows, TitleStyle.Render("🔒 Unlock Wallet"))
	rows = append(rows, "")
	rows = append(rows, SubtitleStyle.Render("Enter your password to unlock the wallet"))
	rows = append(rows, "")

	// Password field (masked)
	passwordMask := ""
	for i := 0; i < len(v.password); i++ {
		passwordMask += "•"
	}
	rows = append(rows, InputFocusedStyle.Width(40).Render(passwordMask))
	rows = append(rows, "")

	// Error message
	if v.err != nil {
		rows = append(rows, ErrorStyle.Render(v.err.Error()))
		rows = append(rows, "")
	}

	// Failed attempts warning
	if v.failedAttempts > 0 {
		attemptsLeft := 3 - v.failedAttempts
		rows = append(rows, WarningStyle.Render(fmt.Sprintf("⚠ %d attempts remaining", attemptsLeft)))
		rows = append(rows, "")
	}

	rows = append(rows, InfoStyle.Render("enter submit • esc quit"))

	return PanelStyle.Render(lipgloss.JoinVertical(lipgloss.Left, rows...))
}

// renderCreate renders the wallet creation screen
func (v *AuthView) renderCreate() string {
	var rows []string
	rows = append(rows, TitleStyle.Render("🆕 Create New Wallet"))
	rows = append(rows, "")
	rows = append(rows, SubtitleStyle.Render("Choose a strong password to encrypt your wallet"))
	rows = append(rows, "")

	// Password field (masked)
	passwordMask := ""
	for i := 0; i < len(v.password); i++ {
		passwordMask += "•"
	}
	rows = append(rows, InputFocusedStyle.Width(40).Render(passwordMask))
	rows = append(rows, "")

	// Password strength indicator
	strength := v.getPasswordStrength()
	strengthColor := ColorError
	if strength >= 8 {
		strengthColor = ColorSuccess
	} else if strength >= 6 {
		strengthColor = ColorWarning
	}
	rows = append(rows, lipgloss.NewStyle().Foreground(strengthColor).Render(fmt.Sprintf("Strength: %d/10", strength)))
	rows = append(rows, "")

	// Error message
	if v.err != nil {
		rows = append(rows, ErrorStyle.Render(v.err.Error()))
		rows = append(rows, "")
	}

	rows = append(rows, InfoStyle.Render("enter submit • esc cancel"))

	return PanelStyle.Render(lipgloss.JoinVertical(lipgloss.Left, rows...))
}

// renderRestore renders the wallet restoration screen
func (v *AuthView) renderRestore() string {
	var rows []string
	rows = append(rows, TitleStyle.Render("🔄 Restore Wallet"))
	rows = append(rows, "")
	rows = append(rows, SubtitleStyle.Render("Enter a password to encrypt the restored wallet"))
	rows = append(rows, "")

	// Password field (masked)
	passwordMask := ""
	for i := 0; i < len(v.password); i++ {
		passwordMask += "•"
	}
	rows = append(rows, InputFocusedStyle.Width(40).Render(passwordMask))
	rows = append(rows, "")

	// Error message
	if v.err != nil {
		rows = append(rows, ErrorStyle.Render(v.err.Error()))
		rows = append(rows, "")
	}

	rows = append(rows, InfoStyle.Render("enter submit • esc cancel"))

	return PanelStyle.Render(lipgloss.JoinVertical(lipgloss.Left, rows...))
}

// renderLocked renders the locked screen
func (v *AuthView) renderLocked(secondsRemaining int) string {
	var rows []string
	rows = append(rows, TitleStyle.Render("🔒 Wallet Locked"))
	rows = append(rows, "")
	rows = append(rows, ErrorStyle.Render("Too many failed authentication attempts"))
	rows = append(rows, "")
	rows = append(rows, WarningStyle.Render(fmt.Sprintf("Please wait %d seconds before trying again", secondsRemaining)))

	return PanelStyle.Render(lipgloss.JoinVertical(lipgloss.Left, rows...))
}

// authenticate attempts to authenticate with the entered password
func (v *AuthView) authenticate() tea.Cmd {
	return func() tea.Msg {
		// Validate password length
		if len(v.password) < 8 {
			return authResultMsg{
				success: false,
				error:   "password must be at least 8 characters",
			}
		}

		// In a real implementation, this would verify the password
		// For now, we'll just return success
		return authResultMsg{
			success: true,
		}
	}
}

// getPasswordStrength calculates password strength (0-10)
func (v *AuthView) getPasswordStrength() int {
	if len(v.password) == 0 {
		return 0
	}

	strength := 0

	// Length
	if len(v.password) >= 8 {
		strength += 2
	}
	if len(v.password) >= 12 {
		strength += 2
	}
	if len(v.password) >= 16 {
		strength += 1
	}

	// Character variety
	hasLower := false
	hasUpper := false
	hasDigit := false
	hasSpecial := false

	for _, c := range v.password {
		if c >= 'a' && c <= 'z' {
			hasLower = true
		} else if c >= 'A' && c <= 'Z' {
			hasUpper = true
		} else if c >= '0' && c <= '9' {
			hasDigit = true
		} else {
			hasSpecial = true
		}
	}

	if hasLower {
		strength++
	}
	if hasUpper {
		strength++
	}
	if hasDigit {
		strength++
	}
	if hasSpecial {
		strength += 2
	}

	if strength > 10 {
		strength = 10
	}

	return strength
}

// Messages

type authResultMsg struct {
	success bool
	error   string
}
