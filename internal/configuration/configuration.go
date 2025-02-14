package configuration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/BrandonIrizarry/gator/internal/database"
	"github.com/google/uuid"
)

/** A struct for unmarshalling Gator's current JSON configuration. */
type Config struct {
	DbURL           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

/** A struct for containing all necessary global state. */
type state struct {
	// Gator's current JSON configuration.
	Config *Config

	// The full path to the Gator JSON file.
	ConfigFile string

	// The interface to the database itself.
	db *database.Queries
}

/*
  - An abbreviation for the canonical type signature CLI commands have
    as Go functions.
*/
type cliHandler = func(state, ...string) error
type StateType = state

/** The command registry proper. */
var commandRegistry = make(map[string]cliHandler)

/** Helper to facilitate creating a new state. */
func NewState(configBasename string, dbURL string) (state, error) {
	// Get the user's home directory.
	homeDir, err := os.UserHomeDir()

	if err != nil {
		return state{}, err
	}

	// Open the database connection.
	db, err := sql.Open("postgres", dbURL)

	if err != nil {
		return state{}, err
	}

	// With all the data in place, configure the state.
	state := state{
		ConfigFile: fmt.Sprintf("%s/%s", homeDir, configBasename),
		Config:     &Config{},
		db:         database.New(db),
	}

	return state, nil
}

/*
  - Read the contents of the given state struct's config file into the
    'config' portion of the same struct.
*/
func Read(state state) error {
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
func SetUser(state state, username string) error {
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

func GetCommand(commandName string) (cliHandler, error) {
	fn, ok := commandRegistry[commandName]

	if !ok {
		return nil, fmt.Errorf("Nonexistent command '%s'", commandName)
	}

	return fn, nil
}

/*
  - CLI commands rely on _handler functions_ for their underlying
    implementation. These functions will have a name of the form
    "handlerX", where the X denotes the functionality being
    implemented, for example, "handlerLogin" is the function enabling
    the 'gator login' command.

    Note that the string elements of 'args' are not the original
    command line arguments; rather, they are the intended arguments to
    the command itself (_not_ including the command name).
*/
func handlerLogin(state state, args ...string) error {
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

func handlerRegister(state state, args ...string) error {
	if len(args) == 0 {
		return fmt.Errorf("Missing username argument. Who are you registering?")
	}

	newname := args[0]
	ctx := context.Background()

	// Note that, since uuid.UUID is an alias for [16]byte, its
	// zero-value would be '[16]byte{}' (all zeroes). And so a freshly
	// initialized 'CreateUserParams' struct would have an ID field
	// set to this value.
	//
	// Conversely, an existing database row will have this set to
	// something non-zero, which is what we check for here.
	if user, _ := state.db.GetUser(ctx, newname); user.ID != [16]byte{} {
		return fmt.Errorf("User '%s' is already registered", newname)
	}

	newuser, err := state.db.CreateUser(ctx, database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      args[0],
	})

	if err != nil {
		return err
	}

	if err = SetUser(state, newname); err != nil {
		return err
	}

	fmt.Printf("User '%s' has been created", newname)
	fmt.Printf("%v\n", newuser)

	return nil
}

/** Automatically register all handler functions. */
func init() {
	commandRegistry["login"] = handlerLogin
	commandRegistry["register"] = handlerRegister
}
