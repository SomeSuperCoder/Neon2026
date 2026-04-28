package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/poh-blockchain/cmd/wallet/core"
	walletrpc "github.com/poh-blockchain/cmd/wallet/rpc"
)

// AccountsView represents the accounts management view
type AccountsView struct {
	wallet    *core.Wallet
	rpcClient *walletrpc.RPCClient
	selected  int
	err       error
}

// NewAccountsView creates a new accounts view
func NewAccountsView(wallet *core.Wallet, rpcClient *walletrpc.RPCClient) *AccountsView {
	return &AccountsView{
		wallet:    wallet,
		rpcClient: rpcClient,
		selected:  0,
	}
}

// Init initializes the accounts view
func (v *AccountsView) Init() tea.Cmd {
	return nil
}

// Update handles messages for the accounts view
func (v *AccountsView) Update(msg tea.Msg) (ViewModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if v.selected > 0 {
				v.selected--
			}
		case "down", "j":
			accounts := v.wallet.GetAccounts()
			if v.selected < len(accounts)-1 {
				v.selected++
			}
		case "enter":
			// Set active account
			v.wallet.SetActiveAccount(v.selected)
		}
	}

	return v, nil
}

// View renders the accounts view
func (v *AccountsView) View() string {
	accounts := v.wallet.GetAccounts()

	if len(accounts) == 0 {
		return PanelStyle.Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				TitleStyle.Render("Accounts"),
				"",
				SubtitleStyle.Render("No accounts found"),
				"",
				InfoStyle.Render("Press 'a' to add an account"),
			),
		)
	}

	var rows []string
	rows = append(rows, TitleStyle.Render("Accounts"))
	rows = append(rows, "")

	// Table header
	header := lipgloss.JoinHorizontal(
		lipgloss.Left,
		TableHeaderStyle.Width(40).Render("Address"),
		TableHeaderStyle.Width(20).Render("Balance"),
		TableHeaderStyle.Width(20).Render("Label"),
	)
	rows = append(rows, header)

	// Table rows
	for i, account := range accounts {
		style := TableCellStyle
		if i%2 == 1 {
			style = TableCellAltStyle
		}
		if i == v.selected {
			style = TableRowSelectedStyle
		}

		addr := FormatAddress(account.Address)
		balance := FormatBalance(account.Balance)
		label := account.Label
		if label == "" {
			label = "-"
		}

		row := lipgloss.JoinHorizontal(
			lipgloss.Left,
			style.Width(40).Render(addr),
			style.Width(20).Render(balance),
			style.Width(20).Render(label),
		)
		rows = append(rows, row)
	}

	rows = append(rows, "")
	rows = append(rows, InfoStyle.Render("↑/↓ navigate • enter select • a add account"))

	return PanelStyle.Render(lipgloss.JoinVertical(lipgloss.Left, rows...))
}
