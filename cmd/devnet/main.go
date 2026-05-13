package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	// Define command-line flags
	validators := flag.Int("validators", 3, "Number of validators")
	portStart := flag.Int("port", 8000, "Starting port number")
	rpcPort := flag.Int("rpc-port", 8899, "RPC node port")
	dbDir := flag.String("db-dir", "./devnet-data", "Database directory")
	logDir := flag.String("log-dir", "./logs", "Log directory")
	developerAccount := flag.Bool("developer-account", false, "Create developer/tester account with initial coins")

	// Subcommands
	startCmd := flag.NewFlagSet("start", flag.ExitOnError)
	startValidators := startCmd.Int("validators", 3, "Number of validators")
	startPort := startCmd.Int("port", 8000, "Starting port number")
	startRpcPort := startCmd.Int("rpc-port", 8899, "RPC node port")
	startDbDir := startCmd.String("db-dir", "./devnet-data", "Database directory")
	startLogDir := startCmd.String("log-dir", "./logs", "Log directory")
	startDeveloperAccount := startCmd.Bool("developer-account", false, "Create developer/tester account with initial coins")

	statusCmd := flag.NewFlagSet("status", flag.ExitOnError)

	logsCmd := flag.NewFlagSet("logs", flag.ExitOnError)
	logsValidator := logsCmd.String("validator", "", "Validator ID to show logs for (or 'rpc' for RPC node)")

	stopCmd := flag.NewFlagSet("stop", flag.ExitOnError)

	cleanCmd := flag.NewFlagSet("clean", flag.ExitOnError)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: devnet [OPTIONS] [COMMAND]\n\n")
		fmt.Fprintf(os.Stderr, "Local validator network creation and management\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nCommands:\n")
		fmt.Fprintf(os.Stderr, "  start     Start devnet with specified number of validators\n")
		fmt.Fprintf(os.Stderr, "  status    Show devnet status\n")
		fmt.Fprintf(os.Stderr, "  logs      Show logs for validators or RPC node\n")
		fmt.Fprintf(os.Stderr, "  stop      Stop running devnet\n")
		fmt.Fprintf(os.Stderr, "  clean     Stop devnet and clean all data\n")
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  devnet start --validators 3\n")
		fmt.Fprintf(os.Stderr, "  devnet start --validators 3 --developer-account\n")
		fmt.Fprintf(os.Stderr, "  devnet status\n")
		fmt.Fprintf(os.Stderr, "  devnet logs --validator 1\n")
		fmt.Fprintf(os.Stderr, "  devnet logs --validator rpc\n")
		fmt.Fprintf(os.Stderr, "  devnet stop\n")
		fmt.Fprintf(os.Stderr, "  devnet clean\n")
	}

	flag.Parse()

	// Handle subcommands
	if len(flag.Args()) > 0 {
		switch flag.Args()[0] {
		case "start":
			startCmd.Parse(flag.Args()[1:])
			fmt.Printf("Starting devnet with %d validators\n", *startValidators)
			if *startDeveloperAccount {
				fmt.Println("Creating developer/tester account with initial coins...")
				// TODO: Create developer account and store seed phrase in devnet_tester.txt
			}
			fmt.Printf("Port range: %d-%d\n", *startPort, *startPort+*startValidators-1)
			fmt.Printf("RPC port: %d\n", *startRpcPort)
			fmt.Printf("Database directory: %s\n", *startDbDir)
			fmt.Printf("Log directory: %s\n", *startLogDir)
			// TODO: Implement devnet startup logic
			return
		case "status":
			statusCmd.Parse(flag.Args()[1:])
			fmt.Println("Checking devnet status...")
			// TODO: Implement status check logic
			return
		case "logs":
			logsCmd.Parse(flag.Args()[1:])
			if *logsValidator == "" {
				fmt.Println("Showing logs for all validators and RPC node...")
				// TODO: Show all logs
			} else if *logsValidator == "rpc" {
				fmt.Println("Showing logs for RPC node...")
				// TODO: Show RPC logs
			} else {
				fmt.Printf("Showing logs for validator %s...\n", *logsValidator)
				// TODO: Show specific validator logs
			}
			return
		case "stop":
			stopCmd.Parse(flag.Args()[1:])
			fmt.Println("Stopping devnet...")
			// TODO: Implement stop logic
			return
		case "clean":
			cleanCmd.Parse(flag.Args()[1:])
			fmt.Println("Cleaning devnet data...")
			// TODO: Implement cleanup logic
			return
		default:
			fmt.Fprintf(os.Stderr, "Unknown command: %s\n", flag.Args()[0])
			flag.Usage()
			os.Exit(1)
		}
	}

	// Default mode: start devnet with default options
	fmt.Printf("Starting devnet with %d validators\n", *validators)
	if *developerAccount {
		fmt.Println("Creating developer/tester account with initial coins...")
		// TODO: Create developer account and store seed phrase in devnet_tester.txt
	}
	fmt.Printf("Port range: %d-%d\n", *portStart, *portStart+*validators-1)
	fmt.Printf("RPC port: %d\n", *rpcPort)
	fmt.Printf("Database directory: %s\n", *dbDir)
	fmt.Printf("Log directory: %s\n", *logDir)
	// TODO: Implement devnet startup logic
}
