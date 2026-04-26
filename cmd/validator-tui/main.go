package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dgraph-io/badger/v4"
	"github.com/poh-blockchain/internal/filestore"
	"github.com/poh-blockchain/internal/genesis"
	"github.com/poh-blockchain/internal/quanticscript"
)

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			Background(lipgloss.Color("#1a1a1a")).
			Padding(0, 1).
			MarginBottom(1)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	activeValidatorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#04B575"))

	inactiveValidatorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF6B6B"))

	slashedValidatorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF0000")).
				Bold(true)

	tableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FAFAFA")).
				Background(lipgloss.Color("#5A5A5A")).
				Padding(0, 1)

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			MarginTop(1)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true)
)

// FileType represents the classification of a file by its prefix pattern
type FileType int

const (
	FileTypeUnknown FileType = iota
	FileTypeValidatorRecord
	FileTypeStakeAccount
	FileTypeEpochState
	FileTypeRewardPool
)

// ClassifyFile determines the type of a file based on its ID and data
func ClassifyFile(fileID filestore.FileID, file *filestore.File) FileType {
	if fileID == genesis.EpochStateFileID {
		return FileTypeEpochState
	}
	if fileID == genesis.RewardPoolFileID {
		return FileTypeRewardPool
	}

	if len(file.Data) == 66 {
		return FileTypeValidatorRecord
	}

	if len(file.Data) == 89 {
		return FileTypeStakeAccount
	}

	if len(file.Data) >= 32 {
		return FileTypeEpochState
	}

	return FileTypeUnknown
}

// ValidatorInfo holds parsed validator record information
type ValidatorInfo struct {
	FileID              filestore.FileID
	Pubkey              [32]byte
	Commission          int64
	TotalDelegatedStake int64
	Status              uint8
	BlocksProduced      int64
	MissedBlocks        int64
	SlashedThisEpoch    uint8
}

// StakeAccountInfo holds parsed stake account information
type StakeAccountInfo struct {
	FileID            filestore.FileID
	DelegatorPubkey   [32]byte
	ValidatorFileID   filestore.FileID
	StakedAmount      int64
	ActivationEpoch   int64
	Status            uint8
	DeactivationEpoch int64
}

// EpochStateInfo holds parsed epoch state information
type EpochStateInfo struct {
	EpochNumber         int64
	EpochStartSlot      int64
	TotalSlotsInEpoch   int64
	ValidatorSchedule   []filestore.FileID
	MissedBlockCounters []int64
}

// RewardPoolInfo holds parsed reward pool information
type RewardPoolInfo struct {
	Balance              int64
	LastDistributedEpoch int64
}

// ReadOnlyFileStore is a wrapper for read-only FileStore access
type ReadOnlyFileStore struct {
	db *badger.DB
}

// NewReadOnlyFileStore creates a new read-only FileStore wrapper
func NewReadOnlyFileStore(dbPath string) (*ReadOnlyFileStore, error) {
	opts := badger.DefaultOptions(dbPath)
	opts.ReadOnly = true
	opts.Logger = nil

	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open FileStore in read-only mode at %s: %w", dbPath, err)
	}

	return &ReadOnlyFileStore{db: db}, nil
}

// GetFile retrieves a file by ID
func (rofs *ReadOnlyFileStore) GetFile(id filestore.FileID) (*filestore.File, error) {
	var fileData []byte
	err := rofs.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(id[:])
		if err != nil {
			return err
		}
		fileData, err = item.ValueCopy(nil)
		return err
	})

	if err == badger.ErrKeyNotFound {
		return nil, fmt.Errorf("file not found: %s", id.String())
	}
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve file: %w", err)
	}

	file, err := filestore.UnmarshalFile(fileData)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal file: %w", err)
	}

	return file, nil
}

// GetAllFileIDs returns all file IDs in the store
func (rofs *ReadOnlyFileStore) GetAllFileIDs() ([]filestore.FileID, error) {
	var fileIDs []filestore.FileID

	err := rofs.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			key := item.Key()

			var fid filestore.FileID
			if len(key) == 32 {
				copy(fid[:], key)
				fileIDs = append(fileIDs, fid)
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to retrieve file IDs: %w", err)
	}

	return fileIDs, nil
}

// Close closes the read-only FileStore
func (rofs *ReadOnlyFileStore) Close() error {
	if rofs.db != nil {
		return rofs.db.Close()
	}
	return nil
}

// tickMsg is sent on every tick
type tickMsg time.Time

// model holds the Bubble Tea application state
type model struct {
	rofs              *ReadOnlyFileStore
	validators        map[filestore.FileID]*ValidatorInfo
	stakeAccounts     map[filestore.FileID]*StakeAccountInfo
	epochState        *EpochStateInfo
	rewardPool        *RewardPoolInfo
	lastRefresh       time.Time
	localValidatorID  filestore.FileID
	localValidatorSet bool
	err               error
	quitting          bool
}

// Init initializes the model
func (m model) Init() tea.Cmd {
	return tea.Batch(
		tickCmd(),
		m.refreshState,
	)
}

// tickCmd returns a command that sends a tick message every second
func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// refreshState refreshes the state from the FileStore
func (m *model) refreshState() tea.Msg {
	fileIDs, err := m.rofs.GetAllFileIDs()
	if err != nil {
		return err
	}

	m.validators = make(map[filestore.FileID]*ValidatorInfo)
	m.stakeAccounts = make(map[filestore.FileID]*StakeAccountInfo)
	m.epochState = nil
	m.rewardPool = nil

	for _, fileID := range fileIDs {
		file, err := m.rofs.GetFile(fileID)
		if err != nil {
			continue
		}

		fileType := ClassifyFile(fileID, file)

		switch fileType {
		case FileTypeValidatorRecord:
			_ = m.parseValidatorRecord(fileID, file)

		case FileTypeStakeAccount:
			_ = m.parseStakeAccount(fileID, file)

		case FileTypeEpochState:
			_ = m.parseEpochState(file)

		case FileTypeRewardPool:
			_ = m.parseRewardPool(file)
		}
	}

	m.lastRefresh = time.Now()
	return tickMsg(time.Now())
}

// parseValidatorRecord parses a validator record file
func (m *model) parseValidatorRecord(fileID filestore.FileID, file *filestore.File) error {
	pubkey, commission, totalStake, status, blocksProduced, missedBlocks, slashedThisEpoch, err := quanticscript.DeserializeValidatorRecord(file.Data)
	if err != nil {
		return err
	}

	var pubkeyArray [32]byte
	copy(pubkeyArray[:], pubkey)

	m.validators[fileID] = &ValidatorInfo{
		FileID:              fileID,
		Pubkey:              pubkeyArray,
		Commission:          commission,
		TotalDelegatedStake: totalStake,
		Status:              status,
		BlocksProduced:      blocksProduced,
		MissedBlocks:        missedBlocks,
		SlashedThisEpoch:    slashedThisEpoch,
	}

	return nil
}

// parseStakeAccount parses a stake account file
func (m *model) parseStakeAccount(fileID filestore.FileID, file *filestore.File) error {
	delegatorPubkey, validatorFileID, stakedAmount, activationEpoch, status, deactivationEpoch, err := quanticscript.DeserializeStakeAccount(file.Data)
	if err != nil {
		return err
	}

	var delegatorArray [32]byte
	copy(delegatorArray[:], delegatorPubkey)

	var validatorIDArray filestore.FileID
	copy(validatorIDArray[:], validatorFileID)

	m.stakeAccounts[fileID] = &StakeAccountInfo{
		FileID:            fileID,
		DelegatorPubkey:   delegatorArray,
		ValidatorFileID:   validatorIDArray,
		StakedAmount:      stakedAmount,
		ActivationEpoch:   activationEpoch,
		Status:            status,
		DeactivationEpoch: deactivationEpoch,
	}

	return nil
}

// parseEpochState parses the epoch state file
func (m *model) parseEpochState(file *filestore.File) error {
	epochNumber, epochStartSlot, totalSlotsInEpoch, validatorSchedule, missedBlockCounters, err := quanticscript.DeserializeEpochState(file.Data)
	if err != nil {
		return err
	}

	schedule := make([]filestore.FileID, len(validatorSchedule))
	for i, fileIDBytes := range validatorSchedule {
		copy(schedule[i][:], fileIDBytes[:])
	}

	m.epochState = &EpochStateInfo{
		EpochNumber:         epochNumber,
		EpochStartSlot:      epochStartSlot,
		TotalSlotsInEpoch:   totalSlotsInEpoch,
		ValidatorSchedule:   schedule,
		MissedBlockCounters: missedBlockCounters,
	}

	return nil
}

// parseRewardPool parses the reward pool file
func (m *model) parseRewardPool(file *filestore.File) error {
	balance, lastDistributedEpoch, err := quanticscript.DeserializeRewardPool(file.Data)
	if err != nil {
		return err
	}

	m.rewardPool = &RewardPoolInfo{
		Balance:              balance,
		LastDistributedEpoch: lastDistributedEpoch,
	}

	return nil
}

// Update handles messages
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		}

	case tickMsg:
		return m, tea.Batch(
			tickCmd(),
			m.refreshState,
		)

	case error:
		m.err = msg
		return m, nil
	}

	return m, nil
}

// View renders the UI
func (m model) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	if m.err != nil {
		return errorStyle.Render(fmt.Sprintf("Error: %v\n\nPress 'q' to quit.", m.err))
	}

	s := titleStyle.Render("🔗 VALIDATOR TUI DASHBOARD") + "\n\n"

	// Header panel
	s += headerStyle.Render(fmt.Sprintf(
		" Epoch: %d | Slot: %d | Local Validator: %s | Active Validators: %d | Local Stake: %d electrons ",
		m.GetCurrentEpoch(),
		m.GetCurrentSlot(),
		m.GetLocalValidatorStatus(),
		m.GetActiveValidatorCount(),
		m.GetLocalDelegatedStake(),
	)) + "\n\n"

	// Validator table
	if len(m.validators) > 0 {
		s += tableHeaderStyle.Render(" Validator Records ") + "\n\n"
		s += m.renderValidatorTable()
	} else {
		s += "No validators found. Waiting for data...\n"
	}

	s += "\n"

	// Summary footer
	s += headerStyle.Render(fmt.Sprintf(
		" Total Staked: %d electrons | Reward Pool: %d electrons | Est. APY: %.2f%% ",
		m.GetTotalStakedElectrons(),
		m.GetRewardPoolBalance(),
		m.GetEstimatedAPY(),
	)) + "\n\n"

	s += footerStyle.Render(fmt.Sprintf("Last refresh: %s | Press 'q' to quit", m.lastRefresh.Format("15:04:05")))

	return s
}

// renderValidatorTable renders the validator table
func (m model) renderValidatorTable() string {
	header := fmt.Sprintf("%-18s %-12s %-18s %-8s %-8s %-6s %-8s\n",
		"Pubkey (16 hex)", "Status", "Total Stake", "Comm %", "Blocks", "Miss", "Slashed")

	rows := ""
	for _, v := range m.validators {
		pubkeyHex := hex.EncodeToString(v.Pubkey[:8])
		statusStr := m.getStyledStatus(v)

		row := fmt.Sprintf("%-18s %-12s %18d %8d %8d %6d %8d\n",
			pubkeyHex,
			statusStr,
			v.TotalDelegatedStake,
			v.Commission,
			v.BlocksProduced,
			v.MissedBlocks,
			v.SlashedThisEpoch,
		)
		rows += row
	}

	return header + rows
}

// getStyledStatus returns a styled status string
func (m model) getStyledStatus(v *ValidatorInfo) string {
	if v.SlashedThisEpoch == 1 {
		return slashedValidatorStyle.Render("[slashed]")
	}

	switch v.Status {
	case 0:
		return inactiveValidatorStyle.Render("[inactive]")
	case 1:
		return activeValidatorStyle.Render("[active]")
	case 2:
		return inactiveValidatorStyle.Render("[deregistered]")
	default:
		return "[unknown]"
	}
}

// GetActiveValidatorCount returns the number of active validators
func (m *model) GetActiveValidatorCount() int {
	count := 0
	for _, v := range m.validators {
		if v.Status == 1 {
			count++
		}
	}
	return count
}

// GetTotalStakedElectrons returns the total staked electrons across all validators
func (m *model) GetTotalStakedElectrons() int64 {
	total := int64(0)
	for _, v := range m.validators {
		total += v.TotalDelegatedStake
	}
	return total
}

// GetRewardPoolBalance returns the reward pool balance
func (m *model) GetRewardPoolBalance() int64 {
	if m.rewardPool == nil {
		return 0
	}
	return m.rewardPool.Balance
}

// GetEstimatedAPY calculates estimated APY
func (m *model) GetEstimatedAPY() float64 {
	totalStaked := m.GetTotalStakedElectrons()
	if totalStaked == 0 {
		return 0
	}

	rewardsPerEpoch := totalStaked / 100
	epochsPerYear := int64(365 * 24 * 60 * 60 * 1000 / (432000 * 400))

	if epochsPerYear == 0 {
		return 0
	}

	apy := float64(rewardsPerEpoch*epochsPerYear) / float64(totalStaked) * 100
	return apy
}

// GetCurrentSlot calculates the current slot
func (m *model) GetCurrentSlot() int64 {
	if m.epochState == nil {
		return 0
	}
	return m.epochState.EpochStartSlot
}

// GetCurrentEpoch returns the current epoch number
func (m *model) GetCurrentEpoch() int64 {
	if m.epochState == nil {
		return 0
	}
	return m.epochState.EpochNumber
}

// GetLocalValidatorStatus returns the status of the local validator
func (m *model) GetLocalValidatorStatus() string {
	if !m.localValidatorSet {
		return "not_set"
	}

	v, ok := m.validators[m.localValidatorID]
	if !ok {
		return "not_found"
	}

	switch v.Status {
	case 0:
		return "inactive"
	case 1:
		return "active"
	case 2:
		return "deregistered"
	default:
		return "unknown"
	}
}

// GetLocalDelegatedStake returns the total delegated stake for the local validator
func (m *model) GetLocalDelegatedStake() int64 {
	if !m.localValidatorSet {
		return 0
	}

	v, ok := m.validators[m.localValidatorID]
	if !ok {
		return 0
	}

	return v.TotalDelegatedStake
}

// ensureStateDirectory creates the state directory if it doesn't exist
func ensureStateDirectory(statePath string) error {
	// Check if path exists
	info, err := os.Stat(statePath)
	if err == nil {
		// Path exists, check if it's a directory
		if !info.IsDir() {
			return fmt.Errorf("state path exists but is not a directory: %s", statePath)
		}
		return nil
	}

	if !os.IsNotExist(err) {
		return fmt.Errorf("failed to check state path: %w", err)
	}

	// Directory doesn't exist, create it
	fmt.Printf("Creating state directory: %s\n", statePath)
	if err := os.MkdirAll(statePath, 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	// Initialize with genesis state
	fmt.Println("Initializing genesis state...")
	fs, err := filestore.NewFileStore(statePath)
	if err != nil {
		return fmt.Errorf("failed to create FileStore: %w", err)
	}
	defer fs.Close()

	// Load builtin programs with minimal bytecode
	// Note: We use minimal bytecode for system and token programs
	// The staking program is not loaded in this minimal initialization
	systemBytecode := []byte{0x00, 0x01} // Minimal valid bytecode
	tokenBytecode := []byte{0x00, 0x01}  // Minimal valid bytecode

	if err := genesis.LoadBuiltinPrograms(fs, systemBytecode, tokenBytecode, nil); err != nil {
		return fmt.Errorf("failed to load builtin programs: %w", err)
	}

	// Initialize DPoS genesis with a minimal validator set
	// Note: InitializeDPoSGenesis requires at least one validator
	genesisConfig := genesis.GenesisConfig{
		EpochLength: 432000, // Default epoch length
		GenesisValidators: []genesis.GenesisValidator{
			{
				PublicKey:   [32]byte{}, // Placeholder validator
				StakeAmount: 1000000,    // Minimum stake
			},
		},
	}
	if err := genesis.InitializeDPoSGenesis(fs, genesisConfig); err != nil {
		return fmt.Errorf("failed to initialize DPoS genesis: %w", err)
	}

	fmt.Println("Genesis state initialized successfully!")
	return nil
}

// main is the entry point for the validator TUI app
func main() {
	stateFlag := flag.String("state", "", "Path to the FileStore database (required)")
	flag.Parse()

	if *stateFlag == "" {
		fmt.Fprintf(os.Stderr, "Error: --state flag is required\n")
		fmt.Fprintf(os.Stderr, "Usage: validator-tui --state <path_to_filestore>\n")
		os.Exit(1)
	}

	// Resolve absolute path
	statePath, err := filepath.Abs(*stateFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to resolve state path: %v\n", err)
		os.Exit(1)
	}

	// Ensure state directory exists
	if err := ensureStateDirectory(statePath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Open FileStore in read-only mode
	rofs, err := NewReadOnlyFileStore(statePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to open FileStore: %v\n", err)
		os.Exit(1)
	}
	defer rofs.Close()

	// Create initial model
	m := model{
		rofs:          rofs,
		validators:    make(map[filestore.FileID]*ValidatorInfo),
		stakeAccounts: make(map[filestore.FileID]*StakeAccountInfo),
	}

	// Start Bubble Tea program
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
