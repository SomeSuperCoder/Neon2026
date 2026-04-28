package ui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/poh-blockchain/cmd/wallet/core"
	walletrpc "github.com/poh-blockchain/cmd/wallet/rpc"
	"github.com/poh-blockchain/internal/rpc"
)

// DashboardView represents the main dashboard view
type DashboardView struct {
	wallet       *core.Wallet
	rpcClient    *walletrpc.RPCClient
	totalBalance int64
	recentTxs    []rpc.TransactionRecord
	blockHeight  int64
	loading      bool
	err          error
}

// NewDashboardView creates a new dashboard view
func NewDashboardView(wallet *core.Wallet, rpcClient *walletrpc.RPCClient) *DashboardView {
	return &DashboardView{
		wallet:    wallet,
		rpcClient: rpcClient,
		loading:   true,
	}
}

// Init initializes the dashboard view
func (v *DashboardView) Init() tea.Cmd {
	return tea.Batch(
		v.fetchData(),
		v.tickCmd(),
	)
}

// Update handles messages for the dashboard view
func (v *DashboardView) Update(msg tea.Msg) (ViewModel, tea.Cmd) {
	switch msg := msg.(type) {
	case dashboardDataMsg:
		v.totalBalance = msg.totalBalance
		v.recentTxs = msg.recentTxs
		v.blockHeight = msg.blockHeight
		v.loading = false
		v.err = msg.err
		return v, nil

	case tickMsg:
		return v, tea.Batch(v.fetchData(), v.tickCmd())
	}

	return v, nil
}

// View renders the dashboard view
func (v *DashboardView) View() string {
	if v.loading {
		return v.renderLoading()
	}

	if v.err != nil {
		return v.renderError()
	}

	return v.renderDashboard()
}

// renderLoading renders the loading state
func (v *DashboardView) renderLoading() string {
	spinner := "⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏"
	frame := int(time.Now().UnixNano()/100000000) % len(spinner)

	return PanelStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			TitleStyle.Render("Dashboard"),
			"",
			lipgloss.NewStyle().Foreground(ColorPrimary).Render(
				fmt.Sprintf("%c Loading...", rune(spinner[frame])),
			),
		),
	)
}

// renderError renders the error state
func (v *DashboardView) renderError() string {
	return PanelStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			TitleStyle.Render("Dashboard"),
			"",
			ErrorStyle.Render("Error: "+v.err.Error()),
			"",
			InfoStyle.Render("Press 'r' to retry"),
		),
	)
}

// renderDashboard renders the main dashboard content
func (v *DashboardView) renderDashboard() string {
	// Balance section
	balanceSection := v.renderBalanceSection()

	// Recent transactions section
	txSection := v.renderRecentTransactions()

	// Network info section
	networkSection := v.renderNetworkInfo()

	return PanelStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			TitleStyle.Render("Dashboard"),
			"",
			balanceSection,
			"",
			txSection,
			"",
			networkSection,
		),
	)
}

// renderBalanceSection renders the total balance
func (v *DashboardView) renderBalanceSection() string {
	label := lipgloss.NewStyle().Foreground(ColorTextDim).Render("Total Balance:")
	balance := FormatBalance(v.totalBalance)

	return lipgloss.JoinHorizontal(lipgloss.Left, label, " ", balance, " Neon")
}

// renderRecentTransactions renders recent transactions
func (v *DashboardView) renderRecentTransactions() string {
	title := lipgloss.NewStyle().
		Foreground(ColorPrimary).
		Bold(true).
		Render("Recent Transactions")

	if len(v.recentTxs) == 0 {
		return lipgloss.JoinVertical(
			lipgloss.Left,
			title,
			"",
			SubtitleStyle.Render("No recent transactions"),
		)
	}

	var txLines []string
	txLines = append(txLines, title)
	txLines = append(txLines, "")

	// Show up to 5 recent transactions
	count := len(v.recentTxs)
	if count > 5 {
		count = 5
	}

	for i := 0; i < count; i++ {
		tx := v.recentTxs[i]
		txLines = append(txLines, v.formatTransaction(tx))
	}

	return lipgloss.JoinVertical(lipgloss.Left, txLines...)
}

// formatTransaction formats a single transaction for display
func (v *DashboardView) formatTransaction(tx rpc.TransactionRecord) string {
	// Determine if incoming or outgoing
	activeAccount := v.wallet.GetActiveAccount()
	incoming := false
	var amount int64
	var counterparty string

	// Parse transaction to determine direction and amount
	// This is simplified - in reality we'd need to parse instructions
	if len(tx.Instructions) > 0 {
		inst := tx.Instructions[0]
		if len(inst.Accounts) >= 2 {
			if inst.Accounts[1] == activeAccount.Address {
				incoming = true
			}
			if incoming {
				counterparty = inst.Accounts[0]
			} else {
				counterparty = inst.Accounts[1]
			}
		}
		// Amount would be parsed from instruction data
		amount = 1000 // Placeholder
	}

	direction := FormatAmount(amount, incoming)
	addr := FormatAddress(counterparty)

	return lipgloss.JoinHorizontal(lipgloss.Left, direction, " ", addr)
}

// renderNetworkInfo renders network information
func (v *DashboardView) renderNetworkInfo() string {
	heightLabel := lipgloss.NewStyle().Foreground(ColorTextDim).Render("Block Height:")
	height := lipgloss.NewStyle().Foreground(ColorText).Render(fmt.Sprintf("%d", v.blockHeight))

	networkLabel := lipgloss.NewStyle().Foreground(ColorTextDim).Render("Network:")
	network := lipgloss.NewStyle().Foreground(ColorAccent).Render("Devnet")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Left, heightLabel, " ", height),
		lipgloss.JoinHorizontal(lipgloss.Left, networkLabel, " ", network),
	)
}

// fetchData fetches dashboard data from RPC
func (v *DashboardView) fetchData() tea.Cmd {
	return func() tea.Msg {
		var totalBalance int64
		var recentTxs []rpc.TransactionRecord
		var blockHeight int64
		var err error

		// Get block height
		blockHeight, err = v.rpcClient.GetBlockHeight()
		if err != nil {
			return dashboardDataMsg{err: err}
		}

		// Calculate total balance across all accounts
		accounts := v.wallet.GetAccounts()
		for _, account := range accounts {
			balance, balErr := v.rpcClient.GetBalance(account.Address)
			if balErr == nil {
				totalBalance += balance
			}
		}

		// Get recent transactions for active account
		activeAccount := v.wallet.GetActiveAccount()
		if activeAccount != nil {
			recentTxs, err = v.rpcClient.GetTransactionHistory(activeAccount.Address, 5)
			if err != nil {
				// Don't fail if we can't get transactions
				recentTxs = []rpc.TransactionRecord{}
			}
		}

		return dashboardDataMsg{
			totalBalance: totalBalance,
			recentTxs:    recentTxs,
			blockHeight:  blockHeight,
			err:          nil,
		}
	}
}

// tickCmd returns a command that ticks every 5 seconds
func (v *DashboardView) tickCmd() tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Messages

type dashboardDataMsg struct {
	totalBalance int64
	recentTxs    []rpc.TransactionRecord
	blockHeight  int64
	err          error
}

type tickMsg time.Time
