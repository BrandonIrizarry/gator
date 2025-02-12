package main

import (
	"fmt"
	"github.com/BrandonIrizarry/gator/internal/configuration"
	"os"
)

const configBasename = ".gatorconfig.json"

func main() {
	homeDir, err := os.UserHomeDir()

	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
	}

	state := configuration.State{
		ConfigFile: fmt.Sprintf("%s/%s", homeDir, configBasename),
	}

	if err := configuration.Read(&state); err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
	}

	configuration.SetUser(&state, "ted")

	if err := configuration.Read(&state); err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
	}

	fmt.Printf("%v\n%v\n", *state.Config, state.ConfigFile)
}
