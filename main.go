package main

import (
	"fmt"
	"github.com/BrandonIrizarry/gator/internal/configuration"
	"os"
)

const configBasename = ".gatorconfig.json"

func main() {
	// Initialize a new State.
	state, err := configuration.NewState(configBasename)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error defining State: %v\n", err)
		os.Exit(1)
	}

	// Read the current JSON configuration into the State.
	if err := configuration.Read(&state); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	// Parse the current command, and check if everything is OK.
	if len(os.Args) <= 1 {
		fmt.Fprintf(os.Stderr, "No arguments provided\n")
		os.Exit(1)
	}

	commandName := os.Args[1]
	command, err := configuration.GetCommand(commandName)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	// Invoke the given command.
	if err = command(&state, os.Args[2:]...); err != nil {
		fmt.Fprintf(os.Stderr, "In command '%s': %v\n", commandName, err)
		os.Exit(1)
	}

}
