package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/poh-blockchain/cmd/wallet/core"
)

// WizardView represents the wallet creation/restoration wizard
type WizardView struct {
	wallet      *core.Wallet
	step        WizardStep
	wordCount   int // 12 or 24
	seedPhrase  string
	password    string
	confirmPass string
	focusIndex  int
	err         error
	result      *WizardResult
}

// WizardStep represents the current step in the wizard
type WizardStep int

const (
	StepChooseAction WizardStep = iota
	StepChooseWordCount
	StepDisplaySeedPhrase
	StepConfirmSeedPhrase
	StepEnterSeedPhrase
	StepSetPassword
	StepConfirmPassword
	StepComplete
)

// WizardResult holds the result of the wizard
type WizardResult struct {
	Success    bool
	SeedPhrase string
	Password   string
	Error      string
}

// NewWizardView creates a new wizard view
func NewWizardView(wallet *core.Wallet) *WizardView {
	return &WizardView{
		wallet:     wallet,
		step:       StepChooseAction,
		wordCount:  12,
		focusIndex: 0,
	}
}

// Init initializes the wizard view
func (v *WizardView) Init() tea.Cmd {
	return nil
}

// Update handles messages for the wizard view
func (v *WizardView) Update(msg tea.Msg) (ViewModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return v.handleKeyPress(msg)

	case seedPhraseGeneratedMsg:
		v.seedPhrase = msg.seedPhrase
		v.step = StepDisplaySeedPhrase
		return v, nil

	case wizardCompleteMsg:
		v.result = &WizardResult{
			Success:    msg.success,
			SeedPhrase: v.seedPhrase,
			Password:   v.password,
			Error:      msg.error,
		}
		v.step = StepComplete
		return v, nil
	}

	return v, nil
}

// View renders the wizard view
func (v *WizardView) View() string {
	switch v.step {
	case StepChooseAction:
		return v.renderChooseAction()
	case StepChooseWordCount:
		return v.renderChooseWordCount()
	case StepDisplaySeedPhrase:
		return v.renderDisplaySeedPhrase()
	case StepConfirmSeedPhrase:
		return v.renderConfirmSeedPhrase()
	case StepEnterSeedPhrase:
		return v.renderEnterSeedPhrase()
	case StepSetPassword:
		return v.renderSetPassword()
	case StepConfirmPassword:
		return v.renderConfirmPassword()
	case StepComplete:
		return v.renderComplete()
	default:
		return v.renderChooseAction()
	}
}

// renderChooseAction renders the initial action selection
func (v *WizardView) renderChooseAction() string {
	var rows []string
	rows = append(rows, TitleStyle.Render("Welcome to Neon Wallet"))
	rows = append(rows, "")
	rows = append(rows, SubtitleStyle.Render("Choose an option:"))
	rows = append(rows, "")

	option1Style := ButtonStyle
	option2Style := ButtonStyle
	if v.focusIndex == 0 {
		option1Style = ButtonActiveStyle
	} else {
		option2Style = ButtonActiveStyle
	}

	rows = append(rows, option1Style.Render("1. Create New Wallet"))
	rows = append(rows, option2Style.Render("2. Restore from Seed Phrase"))
	rows = append(rows, "")
	rows = append(rows, InfoStyle.Render("↑/↓ navigate • enter select • q quit"))

	return PanelStyle.Render(lipgloss.JoinVertical(lipgloss.Left, rows...))
}

// renderChooseWordCount renders the word count selection
func (v *WizardView) renderChooseWordCount() string {
	var rows []string
	rows = append(rows, TitleStyle.Render("Choose Seed Phrase Length"))
	rows = append(rows, "")
	rows = append(rows, SubtitleStyle.Render("Select the number of words for your seed phrase:"))
	rows = append(rows, "")

	option1Style := ButtonStyle
	option2Style := ButtonStyle
	if v.focusIndex == 0 {
		option1Style = ButtonActiveStyle
	} else {
		option2Style = ButtonActiveStyle
	}

	rows = append(rows, option1Style.Render("1. 12 words (recommended)"))
	rows = append(rows, option2Style.Render("2. 24 words (more secure)"))
	rows = append(rows, "")
	rows = append(rows, InfoStyle.Render("↑/↓ navigate • enter select • esc back"))

	return PanelStyle.Render(lipgloss.JoinVertical(lipgloss.Left, rows...))
}

// renderDisplaySeedPhrase renders the seed phrase display
func (v *WizardView) renderDisplaySeedPhrase() string {
	var rows []string
	rows = append(rows, TitleStyle.Render("⚠️  Your Seed Phrase"))
	rows = append(rows, "")
	rows = append(rows, WarningStyle.Render("IMPORTANT: Write down these words in order and keep them safe!"))
	rows = append(rows, SubtitleStyle.Render("This is the ONLY time you will see your seed phrase."))
	rows = append(rows, "")

	// Display seed phrase words in a grid
	words := strings.Split(v.seedPhrase, " ")
	for i := 0; i < len(words); i += 3 {
		var line []string
		for j := 0; j < 3 && i+j < len(words); j++ {
			wordNum := fmt.Sprintf("%2d.", i+j+1)
			word := words[i+j]
			line = append(line, lipgloss.NewStyle().
				Foreground(ColorAccent).
				Width(20).
				Render(fmt.Sprintf("%s %s", wordNum, word)))
		}
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Left, line...))
	}

	rows = append(rows, "")
	rows = append(rows, ErrorStyle.Render("⚠️  Never share your seed phrase with anyone!"))
	rows = append(rows, "")
	rows = append(rows, InfoStyle.Render("Press enter when you have written down your seed phrase"))

	return PanelStyle.Render(lipgloss.JoinVertical(lipgloss.Left, rows...))
}

// renderConfirmSeedPhrase renders the seed phrase confirmation
func (v *WizardView) renderConfirmSeedPhrase() string {
	var rows []string
	rows = append(rows, TitleStyle.Render("Confirm Seed Phrase"))
	rows = append(rows, "")
	rows = append(rows, SubtitleStyle.Render("Have you written down your seed phrase?"))
	rows = append(rows, "")

	option1Style := ButtonStyle
	option2Style := ButtonStyle
	if v.focusIndex == 0 {
		option1Style = ButtonActiveStyle
	} else {
		option2Style = ButtonActiveStyle
	}

	rows = append(rows, option1Style.Render("Yes, I have written it down"))
	rows = append(rows, option2Style.Render("No, show it again"))
	rows = append(rows, "")
	rows = append(rows, InfoStyle.Render("↑/↓ navigate • enter select"))

	return PanelStyle.Render(lipgloss.JoinVertical(lipgloss.Left, rows...))
}

// renderEnterSeedPhrase renders the seed phrase entry form
func (v *WizardView) renderEnterSeedPhrase() string {
	var rows []string
	rows = append(rows, TitleStyle.Render("Enter Seed Phrase"))
	rows = append(rows, "")
	rows = append(rows, SubtitleStyle.Render("Enter your seed phrase (space-separated words):"))
	rows = append(rows, "")

	rows = append(rows, InputFocusedStyle.Width(60).Render(v.seedPhrase))
	rows = append(rows, "")

	if v.err != nil {
		rows = append(rows, ErrorStyle.Render(v.err.Error()))
		rows = append(rows, "")
	}

	rows = append(rows, InfoStyle.Render("enter continue • esc back"))

	return PanelStyle.Render(lipgloss.JoinVertical(lipgloss.Left, rows...))
}

// renderSetPassword renders the password entry form
func (v *WizardView) renderSetPassword() string {
	var rows []string
	rows = append(rows, TitleStyle.Render("Set Password"))
	rows = append(rows, "")
	rows = append(rows, SubtitleStyle.Render("Choose a strong password to encrypt your wallet:"))
	rows = append(rows, "")

	// Password field (masked)
	passwordMask := ""
	for i := 0; i < len(v.password); i++ {
		passwordMask += "•"
	}
	rows = append(rows, InputFocusedStyle.Width(40).Render(passwordMask))
	rows = append(rows, "")

	// Password strength
	strength := v.getPasswordStrength()
	strengthColor := ColorError
	if strength >= 8 {
		strengthColor = ColorSuccess
	} else if strength >= 6 {
		strengthColor = ColorWarning
	}
	rows = append(rows, lipgloss.NewStyle().Foreground(strengthColor).Render(fmt.Sprintf("Strength: %d/10", strength)))
	rows = append(rows, "")

	if v.err != nil {
		rows = append(rows, ErrorStyle.Render(v.err.Error()))
		rows = append(rows, "")
	}

	rows = append(rows, InfoStyle.Render("enter continue • esc back"))

	return PanelStyle.Render(lipgloss.JoinVertical(lipgloss.Left, rows...))
}

// renderConfirmPassword renders the password confirmation form
func (v *WizardView) renderConfirmPassword() string {
	var rows []string
	rows = append(rows, TitleStyle.Render("Confirm Password"))
	rows = append(rows, "")
	rows = append(rows, SubtitleStyle.Render("Re-enter your password:"))
	rows = append(rows, "")

	// Password field (masked)
	passwordMask := ""
	for i := 0; i < len(v.confirmPass); i++ {
		passwordMask += "•"
	}
	rows = append(rows, InputFocusedStyle.Width(40).Render(passwordMask))
	rows = append(rows, "")

	if v.err != nil {
		rows = append(rows, ErrorStyle.Render(v.err.Error()))
		rows = append(rows, "")
	}

	rows = append(rows, InfoStyle.Render("enter finish • esc back"))

	return PanelStyle.Render(lipgloss.JoinVertical(lipgloss.Left, rows...))
}

// renderComplete renders the completion screen
func (v *WizardView) renderComplete() string {
	var rows []string

	if v.result.Success {
		rows = append(rows, TitleStyle.Render("✓ Wallet Created Successfully"))
		rows = append(rows, "")
		rows = append(rows, SuccessStyle.Render("Your wallet has been created and encrypted."))
		rows = append(rows, "")
		rows = append(rows, SubtitleStyle.Render("You can now use your wallet to manage your accounts."))
	} else {
		rows = append(rows, TitleStyle.Render("✗ Wallet Creation Failed"))
		rows = append(rows, "")
		rows = append(rows, ErrorStyle.Render(v.result.Error))
	}

	rows = append(rows, "")
	rows = append(rows, InfoStyle.Render("Press enter to continue"))

	return PanelStyle.Render(lipgloss.JoinVertical(lipgloss.Left, rows...))
}

// handleKeyPress handles keyboard input for the wizard
func (v *WizardView) handleKeyPress(msg tea.KeyMsg) (ViewModel, tea.Cmd) {
	switch v.step {
	case StepChooseAction:
		switch msg.String() {
		case "up", "k":
			v.focusIndex = 0
		case "down", "j":
			v.focusIndex = 1
		case "enter":
			if v.focusIndex == 0 {
				v.step = StepChooseWordCount
			} else {
				v.step = StepEnterSeedPhrase
			}
		case "q":
			return v, tea.Quit
		}

	case StepChooseWordCount:
		switch msg.String() {
		case "up", "k":
			v.focusIndex = 0
		case "down", "j":
			v.focusIndex = 1
		case "enter":
			if v.focusIndex == 0 {
				v.wordCount = 12
			} else {
				v.wordCount = 24
			}
			return v, v.generateSeedPhrase()
		case "esc":
			v.step = StepChooseAction
		}

	case StepDisplaySeedPhrase:
		if msg.String() == "enter" {
			v.step = StepConfirmSeedPhrase
		}

	case StepConfirmSeedPhrase:
		switch msg.String() {
		case "up", "k":
			v.focusIndex = 0
		case "down", "j":
			v.focusIndex = 1
		case "enter":
			if v.focusIndex == 0 {
				v.step = StepSetPassword
			} else {
				v.step = StepDisplaySeedPhrase
			}
		}

	case StepEnterSeedPhrase:
		switch msg.String() {
		case "enter":
			if core.ValidateSeedPhrase(v.seedPhrase) {
				v.err = nil
				v.step = StepSetPassword
			} else {
				v.err = fmt.Errorf("invalid seed phrase")
			}
		case "esc":
			v.step = StepChooseAction
		case "backspace":
			if len(v.seedPhrase) > 0 {
				v.seedPhrase = v.seedPhrase[:len(v.seedPhrase)-1]
			}
		default:
			if len(msg.String()) == 1 || msg.String() == "space" {
				if msg.String() == "space" {
					v.seedPhrase += " "
				} else {
					v.seedPhrase += msg.String()
				}
			}
		}

	case StepSetPassword:
		switch msg.String() {
		case "enter":
			if len(v.password) >= 8 {
				v.err = nil
				v.step = StepConfirmPassword
			} else {
				v.err = fmt.Errorf("password must be at least 8 characters")
			}
		case "esc":
			v.step = StepChooseAction
		case "backspace":
			if len(v.password) > 0 {
				v.password = v.password[:len(v.password)-1]
			}
		default:
			if len(msg.String()) == 1 {
				v.password += msg.String()
			}
		}

	case StepConfirmPassword:
		switch msg.String() {
		case "enter":
			if v.password == v.confirmPass {
				v.err = nil
				return v, v.createWallet()
			} else {
				v.err = fmt.Errorf("passwords do not match")
			}
		case "esc":
			v.step = StepSetPassword
			v.confirmPass = ""
		case "backspace":
			if len(v.confirmPass) > 0 {
				v.confirmPass = v.confirmPass[:len(v.confirmPass)-1]
			}
		default:
			if len(msg.String()) == 1 {
				v.confirmPass += msg.String()
			}
		}

	case StepComplete:
		if msg.String() == "enter" {
			// Transition to main app would happen here
			return v, tea.Quit
		}
	}

	return v, nil
}

// generateSeedPhrase generates a new seed phrase
func (v *WizardView) generateSeedPhrase() tea.Cmd {
	return func() tea.Msg {
		seedPhrase, err := core.GenerateSeedPhrase(v.wordCount)
		if err != nil {
			return wizardCompleteMsg{
				success: false,
				error:   err.Error(),
			}
		}
		return seedPhraseGeneratedMsg{seedPhrase: seedPhrase}
	}
}

// createWallet creates the wallet with the seed phrase and password
func (v *WizardView) createWallet() tea.Cmd {
	return func() tea.Msg {
		// Import seed phrase
		_, err := v.wallet.ImportSeedPhrase(v.seedPhrase, "Account 1")
		if err != nil {
			return wizardCompleteMsg{
				success: false,
				error:   err.Error(),
			}
		}

		// Save wallet with encryption
		err = v.wallet.Save(v.password)
		if err != nil {
			return wizardCompleteMsg{
				success: false,
				error:   err.Error(),
			}
		}

		return wizardCompleteMsg{success: true}
	}
}

// getPasswordStrength calculates password strength (0-10)
func (v *WizardView) getPasswordStrength() int {
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

type seedPhraseGeneratedMsg struct {
	seedPhrase string
}

type wizardCompleteMsg struct {
	success bool
	error   string
}
