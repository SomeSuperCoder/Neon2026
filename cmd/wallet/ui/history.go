package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/poh-blockchain/cmd/wallet/core"
	walletrpc "github.com/poh-blockchain/cmd/wallet/rpc"
	"github.com/poh-blockchain/internal/rpc"
)

// HistoryView represents the transaction history view
type HistoryView struct {
	wallet       *core.Wallet
	rpcClient    *walletrpc.RPCClient
	transactions []rpc.TransactionRecord
	page         int
	pageSize     int
	loading      bool
	err          error
}

// NewHistoryView creates a new history view
func NewHistoryView(wallet *core.Wallet, rpcClient *walletrpc.RPCClient) *HistoryView {
	return &HistoryView{
		wallet:    wallet,
		rpcClient: rpcClient,
		page:      0,
		pageSize:  20,
		loading:   true,
	}
}

// Init initializes the history view
func (v *HistoryView) Init() tea.Cmd {
	return v.fetchTransactions()
}

// Update handles messages for the history view
func (v *HistoryView) Update(msg tea.Msg) (ViewModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h":
			if v.page > 0 {
				v.page--
				return v, v.fetchTransactions()
			}
		case "right", "l":
			if len(v.transactions) == v.pageSize {
				v.page++
				return v, v.fetchTransactions()
			}
		case "r":
			return v, v.fetchTransactions()
		}

	case historyDataMsg:
		v.transactions = msg.transactions
		v.loading = false
		v.err = msg.err
	}

	return v, nil
}

// View renders the history view
func (v *HistoryView) View() string {
	if v.loading {
		return v.renderLoading()
	}

	if v.err != nil {
		return v.renderError()
	}

	return v.renderHistory()
}

// renderLoading renders the loading state
func (v *HistoryView) renderLoading() string {
	return PanelStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			TitleStyle.Render("Transaction History"),
			"",
			InfoStyle.Render("Loading..."),
		),
	)
}

// renderError renders the error state
func (v *HistoryView) renderError() string {
	return PanelStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			TitleStyle.Render("Transaction History"),
			"",
			ErrorStyle.Render("Error: "+v.err.Error()),
			"",
			InfoStyle.Render("Press 'r' to retry"),
		),
	)
}

// renderHistory renders the transaction history
func (v *HistoryView) renderHistory() string {
	activeAccount := v.wallet.GetActiveAccount()
	if activeAccount == nil {
		return PanelStyle.Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				TitleStyle.Render("Transaction History"),
				"",
				ErrorStyle.Render("No active account"),
			),
		)
	}

	var rows []string
	rows = append(rows, TitleStyle.Render("Transaction History"))
	rows = append(rows, SubtitleStyle.Render(fmt.Sprintf("Account: %s", FormatAddress(activeAccount.Address))))
	rows = append(rows, "")

	if len(v.transactions) == 0 {
		rows = append(rows, SubtitleStyle.Render("No transactions found"))
		return PanelStyle.Render(lipgloss.JoinVertical(lipgloss.Left, rows...))
	}

	// Table header
	header := lipgloss.JoinHorizontal(
		lipgloss.Left,
		TableHeaderStyle.Width(20).Render("Timestamp"),
		TableHeaderStyle.Width(15).Render("Direction"),
		TableHeaderStyle.Width(40).Render("Counterparty"),
		TableHeaderStyle.Width(15).Render("Amount"),
	)
	rows = append(rows, header)

	// Table rows
	for i, tx := range v.transactions {
		style := TableCellStyle
		if i%2 == 1 {
			style = TableCellAltStyle
		}

		timestamp := FormatTimestamp(tx.Timestamp)

		// Determine direction and counterparty
		incoming := false
		var counterparty string
		var amount int64 = 1000 // Placeholder

		if len(tx.Instructions) > 0 {
			inst := tx.Instructions[0]
			if len(inst.Accounts) >= 2 {
				if inst.Accounts[1] == activeAccount.Address {
					incoming = true
					counterparty = inst.Accounts[0]
				} else {
					counterparty = inst.Accounts[1]
				}
			}
		}

		direction := "↑ Sent"
		if incoming {
			direction = "↓ Received"
		}

		row := lipgloss.JoinHorizontal(
			lipgloss.Left,
			style.Width(20).Render(timestamp),
			style.Width(15).Render(direction),
			style.Width(40).Render(FormatAddress(counterparty)),
			style.Width(15).Render(FormatBalance(amount)),
		)
		rows = append(rows, row)
	}

	// Pagination info
	rows = append(rows, "")
	pageInfo := fmt.Sprintf("Page %d", v.page+1)
	rows = append(rows, SubtitleStyle.Render(pageInfo))
	rows = append(rows, "")
	rows = append(rows, InfoStyle.Render("←/h prev page • →/l next page • r refresh"))

	return PanelStyle.Render(lipgloss.JoinVertical(lipgloss.Left, rows...))
}

// fetchTransactions fetches transaction history from RPC
func (v *HistoryView) fetchTransactions() tea.Cmd {
	v.loading = true

	return func() tea.Msg {
		activeAccount := v.wallet.GetActiveAccount()
		if activeAccount == nil {
			return historyDataMsg{err: fmt.Errorf("no active account")}
		}

		transactions, err := v.rpcClient.GetTransactionHistory(activeAccount.Address, v.pageSize)
		if err != nil {
			return historyDataMsg{err: err}
		}

		return historyDataMsg{
			transactions: transactions,
			err:          nil,
		}
	}
}

// Messages

type historyDataMsg struct {
	transactions []rpc.TransactionRecord
	err          error
}
