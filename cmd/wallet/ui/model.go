package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/poh-blockchain/cmd/wallet/core"
	walletrpc "github.com/poh-blockchain/cmd/wallet/rpc"
)

// ViewType represents different views in the wallet
type ViewType int

const (
	ViewDashboard ViewType = iota
	ViewAccounts
	ViewTransfer
	ViewHistory
	ViewSettings
)

// Model is the main application model
type Model struct {
	wallet       *core.Wallet
	rpcClient    *walletrpc.RPCClient
	currentView  ViewType
	views        map[ViewType]ViewModel
	width        int
	height       int
	locked       bool
	lastActivity time.Time
	err          error
}

// ViewModel is the interface that all views must implement
type ViewModel interface {
	Init() tea.Cmd
	Update(msg tea.Msg) (ViewModel, tea.Cmd)
	View() string
}

// NewModel creates a new application model
func NewModel(wallet *core.Wallet, rpcClient *walletrpc.RPCClient) *Model {
	m := &Model{
		wallet:       wallet,
		rpcClient:    rpcClient,
		currentView:  ViewDashboard,
		views:        make(map[ViewType]ViewModel),
		lastActivity: time.Now(),
	}

	// Initialize views
	m.views[ViewDashboard] = NewDashboardView(wallet, rpcClient)
	m.views[ViewAccounts] = NewAccountsView(wallet, rpcClient)
	m.views[ViewTransfer] = NewTransferView(wallet, rpcClient)
	m.views[ViewHistory] = NewHistoryView(wallet, rpcClient)
	m.views[ViewSettings] = NewSettingsView(wallet)

	return m
}

// Init initializes the model
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		m.views[m.currentView].Init(),
	)
}

// Update handles messages and updates the model
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.lastActivity = time.Now()

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "1":
			m.currentView = ViewDashboard
			return m, m.views[m.currentView].Init()
		case "2":
			m.currentView = ViewAccounts
			return m, m.views[m.currentView].Init()
		case "3":
			m.currentView = ViewTransfer
			return m, m.views[m.currentView].Init()
		case "4":
			m.currentView = ViewHistory
			return m, m.views[m.currentView].Init()
		case "5":
			m.currentView = ViewSettings
			return m, m.views[m.currentView].Init()
		}
	}

	// Update current view
	var cmd tea.Cmd
	m.views[m.currentView], cmd = m.views[m.currentView].Update(msg)
	return m, cmd
}

// View renders the model
func (m *Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Render header
	header := m.renderHeader()

	// Render sidebar and content
	sidebar := m.renderSidebar()
	content := m.views[m.currentView].View()

	// Render footer
	footer := m.renderFooter()

	// Combine all parts
	return header + "\n" + sidebar + content + "\n" + footer
}
