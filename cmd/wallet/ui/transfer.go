package ui

import (
	"fmt"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/poh-blockchain/cmd/wallet/core"
	walletrpc "github.com/poh-blockchain/cmd/wallet/rpc"
)

// TransferView represents the transfer view
type TransferView struct {
	wallet        *core.Wallet
	rpcClient     *walletrpc.RPCClient
	recipient     string
	amount        string
	memo          string
	focusIndex    int
	confirmScreen bool
	result        *TransferResult
	submitting    bool
	err           error
}

// TransferResult holds the result of a transfer
type TransferResult struct {
	Success   bool
	Signature string
	Error     string
}

// NewTransferView creates a new transfer view
func NewTransferView(wallet *core.Wallet, rpcClient *walletrpc.RPCClient) *TransferView {
	return &TransferView{
		wallet:     wallet,
		rpcClient:  rpcClient,
		focusIndex: 0,
	}
}

// Init initializes the transfer view
func (v *TransferView) Init() tea.Cmd {
	return nil
}

// Update handles messages for the transfer view
func (v *TransferView) Update(msg tea.Msg) (ViewModel, tea.Cmd) {
	// Handle result screen
	if v.result != nil {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.String() == "enter" || msg.String() == "esc" {
				// Reset form
				v.recipient = ""
				v.amount = ""
				v.memo = ""
				v.result = nil
				v.confirmScreen = false
				v.focusIndex = 0
			}
		}
		return v, nil
	}

	// Handle confirmation screen
	if v.confirmScreen {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "y", "enter":
				return v, v.submitTransfer()
			case "n", "esc":
				v.confirmScreen = false
			}
		case transferResultMsg:
			v.result = &TransferResult{
				Success:   msg.success,
				Signature: msg.signature,
				Error:     msg.error,
			}
			v.submitting = false
		}
		return v, nil
	}

	// Handle form input
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "down":
			v.focusIndex = (v.focusIndex + 1) % 3
		case "shift+tab", "up":
			v.focusIndex = (v.focusIndex - 1 + 3) % 3
		case "enter":
			if v.validateForm() {
				v.confirmScreen = true
			}
		case "backspace":
			v.handleBackspace()
		default:
			v.handleInput(msg.String())
		}
	}

	return v, nil
}

// View renders the transfer view
func (v *TransferView) View() string {
	if v.result != nil {
		return v.renderResult()
	}

	if v.confirmScreen {
		return v.renderConfirmation()
	}

	return v.renderForm()
}

// renderForm renders the transfer form
func (v *TransferView) renderForm() string {
	var rows []string
	rows = append(rows, TitleStyle.Render("Transfer"))
	rows = append(rows, "")

	// Active account info
	activeAccount := v.wallet.GetActiveAccount()
	if activeAccount == nil {
		return PanelStyle.Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				TitleStyle.Render("Transfer"),
				"",
				ErrorStyle.Render("No active account"),
			),
		)
	}

	rows = append(rows, SubtitleStyle.Render(fmt.Sprintf("From: %s", FormatAddress(activeAccount.Address))))
	rows = append(rows, SubtitleStyle.Render(fmt.Sprintf("Balance: %s", FormatBalance(activeAccount.Balance))))
	rows = append(rows, "")

	// Recipient field
	recipientLabel := "Recipient Address:"
	if v.focusIndex == 0 {
		recipientLabel = "→ " + recipientLabel
	}
	rows = append(rows, lipgloss.NewStyle().Foreground(ColorTextDim).Render(recipientLabel))
	recipientStyle := InputStyle
	if v.focusIndex == 0 {
		recipientStyle = InputFocusedStyle
	}
	rows = append(rows, recipientStyle.Width(60).Render(v.recipient))
	rows = append(rows, "")

	// Amount field
	amountLabel := "Amount:"
	if v.focusIndex == 1 {
		amountLabel = "→ " + amountLabel
	}
	rows = append(rows, lipgloss.NewStyle().Foreground(ColorTextDim).Render(amountLabel))
	amountStyle := InputStyle
	if v.focusIndex == 1 {
		amountStyle = InputFocusedStyle
	}
	rows = append(rows, amountStyle.Width(30).Render(v.amount))
	rows = append(rows, "")

	// Memo field
	memoLabel := "Memo (optional):"
	if v.focusIndex == 2 {
		memoLabel = "→ " + memoLabel
	}
	rows = append(rows, lipgloss.NewStyle().Foreground(ColorTextDim).Render(memoLabel))
	memoStyle := InputStyle
	if v.focusIndex == 2 {
		memoStyle = InputFocusedStyle
	}
	rows = append(rows, memoStyle.Width(60).Render(v.memo))
	rows = append(rows, "")

	// Validation errors
	if v.err != nil {
		rows = append(rows, ErrorStyle.Render(v.err.Error()))
		rows = append(rows, "")
	}

	rows = append(rows, InfoStyle.Render("tab/↑↓ navigate • enter submit • esc cancel"))

	return PanelStyle.Render(lipgloss.JoinVertical(lipgloss.Left, rows...))
}

// renderConfirmation renders the confirmation screen
func (v *TransferView) renderConfirmation() string {
	if v.submitting {
		return PanelStyle.Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				TitleStyle.Render("Transfer"),
				"",
				InfoStyle.Render("Submitting transaction..."),
			),
		)
	}

	activeAccount := v.wallet.GetActiveAccount()
	amount, _ := strconv.ParseInt(v.amount, 10, 64)

	var rows []string
	rows = append(rows, TitleStyle.Render("Confirm Transfer"))
	rows = append(rows, "")
	rows = append(rows, lipgloss.NewStyle().Foreground(ColorTextDim).Render("From:"))
	rows = append(rows, FormatAddress(activeAccount.Address))
	rows = append(rows, "")
	rows = append(rows, lipgloss.NewStyle().Foreground(ColorTextDim).Render("To:"))
	rows = append(rows, FormatAddress(v.recipient))
	rows = append(rows, "")
	rows = append(rows, lipgloss.NewStyle().Foreground(ColorTextDim).Render("Amount:"))
	rows = append(rows, FormatBalance(amount))
	rows = append(rows, "")

	if v.memo != "" {
		rows = append(rows, lipgloss.NewStyle().Foreground(ColorTextDim).Render("Memo:"))
		rows = append(rows, lipgloss.NewStyle().Foreground(ColorText).Render(v.memo))
		rows = append(rows, "")
	}

	rows = append(rows, WarningStyle.Render("Confirm this transfer?"))
	rows = append(rows, "")
	rows = append(rows, InfoStyle.Render("y/enter confirm • n/esc cancel"))

	return PanelStyle.Render(lipgloss.JoinVertical(lipgloss.Left, rows...))
}

// renderResult renders the transfer result
func (v *TransferView) renderResult() string {
	var rows []string

	if v.result.Success {
		rows = append(rows, TitleStyle.Render("Transfer Successful"))
		rows = append(rows, "")
		rows = append(rows, SuccessStyle.Render("✓ Transaction submitted"))
		rows = append(rows, "")
		rows = append(rows, lipgloss.NewStyle().Foreground(ColorTextDim).Render("Signature:"))
		rows = append(rows, FormatSignature(v.result.Signature))
	} else {
		rows = append(rows, TitleStyle.Render("Transfer Failed"))
		rows = append(rows, "")
		rows = append(rows, ErrorStyle.Render("✗ "+v.result.Error))
	}

	rows = append(rows, "")
	rows = append(rows, InfoStyle.Render("Press enter to continue"))

	return PanelStyle.Render(lipgloss.JoinVertical(lipgloss.Left, rows...))
}

// handleInput handles character input
func (v *TransferView) handleInput(s string) {
	if len(s) != 1 {
		return
	}

	switch v.focusIndex {
	case 0:
		v.recipient += s
	case 1:
		// Only allow digits for amount
		if s >= "0" && s <= "9" {
			v.amount += s
		}
	case 2:
		v.memo += s
	}
}

// handleBackspace handles backspace key
func (v *TransferView) handleBackspace() {
	switch v.focusIndex {
	case 0:
		if len(v.recipient) > 0 {
			v.recipient = v.recipient[:len(v.recipient)-1]
		}
	case 1:
		if len(v.amount) > 0 {
			v.amount = v.amount[:len(v.amount)-1]
		}
	case 2:
		if len(v.memo) > 0 {
			v.memo = v.memo[:len(v.memo)-1]
		}
	}
}

// validateForm validates the transfer form
func (v *TransferView) validateForm() bool {
	if v.recipient == "" {
		v.err = fmt.Errorf("recipient address is required")
		return false
	}

	if v.amount == "" {
		v.err = fmt.Errorf("amount is required")
		return false
	}

	amount, err := strconv.ParseInt(v.amount, 10, 64)
	if err != nil || amount <= 0 {
		v.err = fmt.Errorf("invalid amount")
		return false
	}

	activeAccount := v.wallet.GetActiveAccount()
	if amount > activeAccount.Balance {
		v.err = fmt.Errorf("insufficient balance")
		return false
	}

	v.err = nil
	return true
}

// submitTransfer submits the transfer transaction
func (v *TransferView) submitTransfer() tea.Cmd {
	v.submitting = true

	return func() tea.Msg {
		amount, _ := strconv.ParseInt(v.amount, 10, 64)

		req := &core.TransferRequest{
			To:     v.recipient,
			Amount: amount,
			Memo:   v.memo,
		}

		signature, err := v.wallet.Transfer(v.rpcClient, req)
		if err != nil {
			return transferResultMsg{
				success: false,
				error:   err.Error(),
			}
		}

		return transferResultMsg{
			success:   true,
			signature: signature,
		}
	}
}

// Messages

type transferResultMsg struct {
	success   bool
	signature string
	error     string
}
