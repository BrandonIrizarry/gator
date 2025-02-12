package configuration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

/** A struct for unmarshalling Gator's current JSON configuration. */
type Config struct {
	DbURL           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

/** A struct for containing all necessary global state. */
type State struct {
	// Gator's current JSON configuration.
	Config *Config

	// The full path to the Gator JSON file.
	ConfigFile string
}

/*
  - An abbreviation for the canonical type signature CLI commands have
    as Go functions.
*/
type cliHandler = func(*State, ...string) error

/** The command registry proper. */
var commandRegistry = make(map[string]cliHandler)

/*
  - Read the contents of the given State struct's config file into the
    'config' portion of the same struct.
*/
func Read(state *State) error {
	if state.ConfigFile == "" {
		return fmt.Errorf("Unconfigured file path to JSON data")
	}

	file, err := os.Open(state.ConfigFile)

	if err != nil {
		return err
	}

	defer file.Close()

	decoder := json.NewDecoder(file)

	if err = decoder.Decode(&state.Config); err != nil {
		return err
	}

	return nil
}

// Set the username in the configuration.
func SetUser(state *State, username string) error {
	if state.ConfigFile == "" {
		return fmt.Errorf("Unconfigured file path to JSON data")
	}

	state.Config.CurrentUserName = username
	buffer := new(bytes.Buffer)

	encoder := json.NewEncoder(buffer)

	if err := encoder.Encode(state.Config); err != nil {
		return err
	}

	if err := os.WriteFile(state.ConfigFile, buffer.Bytes(), 0600); err != nil {
		return err
	}

	return nil
}

/** Runs a given command with the provided state, if the command exists. */
func Run(state *State, commandName string) error {
	fn := commandRegistry[commandName]

	if err := fn(state); err != nil {
		return err
	}

	return nil
}

func RegisterAll() {
	commandRegistry["login"] = handlerLogin
}

/*
  - CLI commands rely on _handler functions_ for their underlying
    implementation. These functions will have a name of the form
    "handlerX", where the X denotes the functionality being
    implemented, for example, "handlerLogin" is the function enabling
    the 'gator login' command.
*/
func handlerLogin(state *State, args ...string) error {
	if len(args) == 0 {
		return fmt.Errorf("Missing username argument")
	}

	username := args[0]

	if err := SetUser(state, username); err != nil {
		return err
	}

	fmt.Printf("The user has been set as '%s'\n", username)
	return nil
}
