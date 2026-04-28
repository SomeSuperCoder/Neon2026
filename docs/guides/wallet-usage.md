# Neon Wallet User Guide

## Overview

Neon Wallet is a modern Terminal User Interface (TUI) wallet for the PoH blockchain. It provides a feature-rich interface for managing accounts, viewing transaction history, and executing transfers with proper seed phrase management.

## Installation

Build the wallet from source:

```bash
go build -o bin/neon-wallet ./cmd/wallet
```

## First Time Setup

When you run the wallet for the first time, it will guide you through the setup process:

```bash
./bin/neon-wallet
```

### Create New Wallet

1. Choose "Create New Wallet"
2. Select seed phrase length (12 or 24 words)
3. Write down your seed phrase (this is the ONLY time you'll see it!)
4. Confirm you've written it down
5. Set a strong password to encrypt your wallet
6. Confirm your password

Your wallet will be created at `~/.poh-wallet/wallet.dat` by default.

### Restore from Seed Phrase

1. Choose "Restore from Seed Phrase"
2. Enter your seed phrase (space-separated words)
3. Set a password to encrypt your wallet
4. Confirm your password

## Command-Line Options

```bash
./bin/neon-wallet [options]
```

Options:
- `--wallet-path <path>`: Custom wallet file location (default: `~/.poh-wallet/wallet.dat`)
- `--rpc-url <url>`: RPC endpoint URL (default: `http://localhost:8899`)

Examples:

```bash
# Use custom wallet location
./bin/neon-wallet --wallet-path /path/to/my/wallet.dat

# Connect to different RPC endpoint
./bin/neon-wallet --rpc-url http://192.168.1.100:8899

# Both options
./bin/neon-wallet --wallet-path ./my-wallet.dat --rpc-url http://localhost:9000
```

## Using the Wallet

### Navigation

- **Number keys (1-5)**: Switch between views
  - `1`: Dashboard
  - `2`: Accounts
  - `3`: Transfer
  - `4`: History
  - `5`: Settings
- **Arrow keys / hjkl**: Navigate within views
- **Enter**: Select/Confirm
- **Esc**: Go back/Cancel
- **q / Ctrl+C**: Quit

### Dashboard View

The dashboard shows:
- Total balance across all accounts
- Current block height
- Recent transactions
- Network status

### Accounts View

Manage your accounts:
- View all accounts with addresses and balances
- Add new accounts (derive from seed phrase)
- Set account labels
- Switch active account

### Transfer View

Send tokens:
1. Enter recipient address
2. Enter amount
3. (Optional) Add memo
4. Review confirmation screen
5. Confirm to submit transaction

### History View

View transaction history:
- All transactions for active account
- Incoming/outgoing indicators
- Transaction details (signature, amount, counterparty)
- Pagination (20 transactions per page)

### Settings View

Configure wallet:
- View/change RPC endpoint
- View auto-lock timeout
- View wallet file path

## Security Features

### Password Protection

Your wallet is encrypted with AES-256-GCM using your password. The password is required every time you start the wallet.

### Auto-Lock

The wallet automatically locks after 5 minutes of inactivity. You'll need to enter your password to unlock it.

### Failed Login Attempts

After 3 failed password attempts, the wallet locks for 30 seconds.

### Seed Phrase Security

- Your seed phrase is displayed ONLY ONCE during wallet creation
- Write it down and store it securely
- Never share your seed phrase with anyone
- The wallet never displays your seed phrase or private keys after creation

## Using with Devnet

Start the devnet with RPC node:

```bash
./devnet.sh start 3
```

The RPC node will be available at `http://localhost:8899` by default.

Then start the wallet:

```bash
./bin/neon-wallet
```

## Troubleshooting

### Cannot connect to RPC node

Error: `Cannot connect to RPC node at http://localhost:8899`

Solution:
1. Make sure the RPC node is running
2. Check the RPC endpoint URL with `--rpc-url` flag
3. Verify network connectivity

### Invalid password

Error: `Invalid password`

Solution:
- Make sure you're entering the correct password
- After 3 failed attempts, wait 30 seconds before trying again
- If you've forgotten your password, you'll need to restore from your seed phrase

### Corrupted wallet file

Error: `Failed to parse wallet file`

Solution:
1. If you have your seed phrase, restore the wallet:
   - Delete the corrupted wallet file
   - Run the wallet again and choose "Restore from Seed Phrase"
2. If you don't have your seed phrase, your funds cannot be recovered

## Best Practices

1. **Backup your seed phrase**: Write it down on paper and store it securely
2. **Use a strong password**: At least 12 characters with mixed case, numbers, and symbols
3. **Keep your wallet file secure**: The default location has restricted permissions (0600)
4. **Regular backups**: Keep multiple copies of your seed phrase in secure locations
5. **Verify addresses**: Always double-check recipient addresses before sending
6. **Test with small amounts**: When sending to a new address, test with a small amount first

## Advanced Usage

### Multiple Seed Phrases

The wallet supports importing multiple seed phrases:
1. Go to Accounts view
2. Select "Import Seed Phrase"
3. Enter the seed phrase
4. Set a label for the account

Each seed phrase derives one account at index 0 (path: m/44'/501'/0'/0'/0').

### Custom Wallet Location

For managing multiple wallets:

```bash
# Personal wallet
./bin/neon-wallet --wallet-path ~/.poh-wallet/personal.dat

# Business wallet
./bin/neon-wallet --wallet-path ~/.poh-wallet/business.dat
```

## Support

For issues or questions:
- Check the logs at `~/.poh-wallet/wallet.log`
- Review the RPC API documentation at `docs/reference/rpc-api.md`
- Check the wallet architecture documentation at `docs/reference/wallet-architecture-update.md`
