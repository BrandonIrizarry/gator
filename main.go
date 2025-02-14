package main

import (
	"fmt"
	"github.com/BrandonIrizarry/gator/internal/configuration"
	_ "github.com/lib/pq"
	"os"
)

const (
	configBasename = ".gatorconfig.json"
	dbURL          = "postgres://postgres:boot.dev@localhost:5432/gator?sslmode=disable"
)

func main() {
	// Initialize a new State.
	state, err := configuration.NewState(configBasename, dbURL)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error defining State: %v\n", err)
		os.Exit(1)
	}

	// Read the current JSON configuration into the State.
	if err := configuration.Read(state); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	// Parse and execute the command.
	if err = parseAndExecute(state, os.Args...); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func parseAndExecute(state configuration.StateType, args ...string) error {
	// Parse the current command, and check if everything is OK.
	if len(args) <= 1 {
		fmt.Fprintf(os.Stderr, "No arguments provided\n")
		os.Exit(1)
	}

	commandName := args[1]
	command, err := configuration.GetCommand(commandName)

	if err != nil {
		return err
	}

	// Invoke the given command.
	if err = command(state, os.Args[2:]...); err != nil {
		return err
	}

	return nil
}
