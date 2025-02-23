package configuration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/BrandonIrizarry/gator/internal/database"
	"github.com/BrandonIrizarry/gator/internal/rss"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/michaljemala/pqerror"
	"os"
	"strconv"
	"time"
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
type cliCommand = func(state, []string) error
type cliLoggedInCommand = func(state, []string, database.User) error

type StateType = state

/** The command registry proper. */
var commandRegistry = make(map[string]cliCommand)

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

func GetCommand(commandName string) (cliCommand, error) {
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
func handlerLogin(state state, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("Missing username argument")
	}

	username := args[0]
	ctx := context.Background()

	// Note that, conversely to 'handlerRegister' (which see), we flag
	// the _absence_ of the specified user.
	if user, _ := state.db.GetUser(ctx, username); user.ID == [16]byte{} {
		return fmt.Errorf("Nonexistent user '%s' (use 'register' to create a new user)", username)
	}

	if err := SetUser(state, username); err != nil {
		return err
	}

	fmt.Printf("The user has been set as '%s'\n", username)
	return nil
}

/*
  - Add (that is, register) the specified user to the 'users'
    table.
*/
func handlerRegister(state state, args []string) error {
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

/*
  - Delete all records in the 'users' table. Used for testing purposes
    only.
*/
func handlerReset(state state, args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("The 'reset' command takes no arguments")
	}

	ctx := context.Background()

	if err := state.db.Reset(ctx); err != nil {
		return err
	}

	return nil
}

func handlerUsers(state state, args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("The 'users' command takes no arguments")
	}

	ctx := context.Background()

	users, err := state.db.GetUsers(ctx)

	if err != nil {
		return err
	}

	for _, user := range users {
		maybeCurrent := ""

		if state.Config.CurrentUserName == user.Name {
			maybeCurrent = " (current)"
		}

		fmt.Printf("%s%s\n", user.Name, maybeCurrent)
	}

	return nil
}

func handlerAgg(state state, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("The 'agg' command takes a single time-between-requests argument")
	}

	duration, err := time.ParseDuration(args[0])

	if err != nil {
		return fmt.Errorf("Unable to parse %q as a duration", duration)
	}

	fmt.Printf("Collecting first feed now; afterwards every %s\n\n", duration)

	if err = scrapeFeeds(state); err != nil {
		return err
	}

	// Continuously scrape the most stale feed.
	ticker := time.NewTicker(duration)
	defer ticker.Stop()

	for range ticker.C {
		if err = scrapeFeeds(state); err != nil {
			return err
		}
	}

	return nil
}

func handlerAddFeed(state state, args []string, currentUser database.User) error {
	if len(args) != 2 {
		return fmt.Errorf("The 'addfeed' command takes a NAME and URL argument")
	}

	feedName := args[0]
	URL := args[1]

	feed, err := state.db.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      feedName,
		Url:       URL,
		UserID:    currentUser.ID,
	})

	if err != nil {
		return fmt.Errorf("'CreateFeed' failed for feed '%s', '%s'", feedName, URL)
	}

	fmt.Println(feed)

	// Also create a feed-follow record for 'currentUser'.
	if _, err = state.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    currentUser.ID,
		FeedID:    feed.ID,
	}); err != nil {
		return fmt.Errorf("Failed to create follow record for:\n\tuser %v\n\tand feed %v\n", currentUser, feed)
	}

	return nil
}

func handlerFeeds(state state, args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("The 'feeds' command takes no arguments")
	}

	ctx := context.Background()
	feeds, err := state.db.GetFeeds(ctx)

	if err != nil {
		return fmt.Errorf("'GetField' failed")
	}

	for _, feed := range feeds {
		user, err := state.db.GetUserByID(ctx, feed.UserID)

		if err != nil {
			return fmt.Errorf("Couldn't get user associated with feed %v\n", feed)
		}

		fmt.Printf("%q, added by user %s\n", feed.Name, user.Name)
	}

	return nil
}

func handlerFollow(state state, args []string, currentUser database.User) error {
	if len(args) != 1 {
		return fmt.Errorf("The 'follow' command takes a single URL argument")
	}

	url := args[0]
	feed, err := state.db.GetFeedByURL(context.Background(), url)

	if err != nil {
		return fmt.Errorf("Failed to fetch feed inside 'handlerFollower'")
	}

	feedInfo, err := state.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    currentUser.ID,
		FeedID:    feed.ID,
	})

	if err != nil {
		return fmt.Errorf("Failed to create follow record for:\n\tuser %v\n\tand feed %v\n", currentUser, feed)
	}

	fmt.Printf("Feed name: %q\nUser name: %q\n", feedInfo.Feedname, feedInfo.Username)

	return nil
}

func handlerFollowing(state state, args []string, currentUser database.User) error {
	if len(args) > 0 {
		return fmt.Errorf("The 'following' command takes no arguments")
	}

	feedFollowsInfo, err := state.db.GetFeedFollowsForUser(context.Background(), currentUser.ID)

	if err != nil {
		return fmt.Errorf("Failed to fetch feed-follows info for user %v\n", currentUser)
	}

	for _, info := range feedFollowsInfo {
		fmt.Println(info.Feedname)
	}

	return nil
}

func handlerUnfollow(state state, args []string, currentUser database.User) error {
	if len(args) != 1 {
		return fmt.Errorf("The command 'unfollow' takes a single URL  argument")
	}

	url := args[0]

	if numDeleted, err := state.db.DeleteFeedFollow(context.Background(), database.DeleteFeedFollowParams{
		UserID: currentUser.ID,
		Url:    url,
	}); err != nil {
		return fmt.Errorf("Failed to delete feed-follow with URL %q\n", url)
	} else if numDeleted == 0 {
		return fmt.Errorf("URL %q doesn't exist in the feed-follows record\n", url)
	}

	return nil
}

func handlerBrowse(state state, args []string, currentUser database.User) error {
	// The cast is required because it's being used as a LIMIT
	// parameter for a query.
	var err error
	var limit64 int64 = 2

	if len(args) == 1 {
		limit64, err = strconv.ParseInt(args[0], 10, 32)

		if err != nil {
			return fmt.Errorf("Can't parse %q as an int\n", args[0])
		}
	} else if len(args) > 1 {
		return fmt.Errorf("Too many args")
	}

	limit := int32(limit64)

	fmt.Println(currentUser, limit)
	posts, err := state.db.GetPostsForUser(context.Background(), database.GetPostsForUserParams{
		UserID: currentUser.ID,
		Limit:  limit,
	})

	if err != nil {
		return err
	}

	fmt.Println(posts)
	for _, post := range posts {
		fmt.Println(post)
	}

	return nil
}

func scrapeFeeds(state state) error {
	feedInfo, err := state.db.GetNextFeedToFetch(context.Background())

	if err != nil {
		// For us, the absence of a feed isn't an error.
		if err == sql.ErrNoRows {
			fmt.Println("<no feeds available at this time>")
			return nil
		} else {
			return fmt.Errorf("Failed to fetch feed %v", feedInfo)
		}
	}

	if err = state.db.MarkFeedFetched(context.Background(), feedInfo.ID); err != nil {
		return fmt.Errorf("Failed to mark as fetched: feed %v", feedInfo)
	}

	rssFeed, err := rss.FetchFeed(context.Background(), feedInfo.Url)

	if err != nil {
		return err
	}

	for _, rssItem := range rssFeed.Channel.Item {
		// Parse the provided publication date into a Go time object.
		pubDate, err := parseRawTime(rssItem.PubDate)

		if err != nil {
			return err
		}

		fmt.Println(rssItem.Link)

		// Save the current rssItem to the 'posts' table.
		post, err := state.db.CreatePost(context.Background(), database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Title:       rssItem.Title,
			Url:         rssItem.Link,
			Description: rssItem.Description,
			PublishedAt: pubDate,
			FeedID:      feedInfo.FeedID,
		})

		if err == sql.ErrNoRows {
			fmt.Printf("Added post %v\n", post)
			continue
		} else {
			var pqErr *pq.Error

			if errors.As(err, &pqErr) {
				constraint := pqErr.Constraint

				if !(pqErr.Code == pqerror.UniqueViolation && constraint == "posts_url_key") {
					return err
				}
			}
		}
	}

	return nil
}

/*
Attempt to parse every RFC layout in the time package.
Return the first valid time.Time. If there are none, return an error.
*/
func parseRawTime(timeStr string) (time.Time, error) {
	layouts := []string{
		time.RFC822,
		time.RFC822Z,
		time.RFC850,
		time.RFC1123,
		time.RFC1123Z,
		time.RFC3339,
		time.RFC3339Nano,
	}

	for _, layout := range layouts {
		t, err := time.Parse(layout, timeStr)

		if err == nil {
			return t, nil
		}
	}

	// Construct a zero-time, to return as a degenerate value.
	var zero time.Time
	return zero, fmt.Errorf("Can't get a valid time from %q; maybe add this layout?", timeStr)
}

/*
  - A function to provide post-login commands (cliLoggedInCommand)
    with the currently logged-in user.

    Essentially, this function converts a given cliLoggedInCommand to
    a cliCommand usable from the main package.
*/
func middlewareWrapper(s state, command cliLoggedInCommand) cliCommand {
	currentUser, err := s.db.GetUser(context.Background(), s.Config.CurrentUserName)

	if err != nil {
		// In case of error, the best we can do is return a dummy
		// function which, when invoked, will return the actual error.
		return func(_ state, _ []string) error {
			return fmt.Errorf("Failed to get user inside middleware wrapper function")
		}
	}

	return func(s state, args []string) error {
		return command(s, args, currentUser)
	}
}

func InitMiddleware(s state) {
	commandRegistry["login"] = handlerLogin
	commandRegistry["register"] = handlerRegister
	commandRegistry["reset"] = handlerReset
	commandRegistry["users"] = handlerUsers
	commandRegistry["agg"] = handlerAgg
	commandRegistry["feeds"] = handlerFeeds

	// The following commands are defined in terms of post-login
	// middleware wrapper calls.
	commandRegistry["addfeed"] = middlewareWrapper(s, handlerAddFeed)
	commandRegistry["follow"] = middlewareWrapper(s, handlerFollow)
	commandRegistry["following"] = middlewareWrapper(s, handlerFollowing)
	commandRegistry["unfollow"] = middlewareWrapper(s, handlerUnfollow)
	commandRegistry["browse"] = middlewareWrapper(s, handlerBrowse)
}
