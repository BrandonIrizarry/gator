# Gator: An RSS Feed Aggregator

Scrape RSS posts from your favorite feeds, and store them locally in a
PostgreSQL database for offline browsing.

Multiple users are allowed and expected to have accounts for browsing
RSS feeds.

## System Requirements and Installation

### System Requirements
This program requires PostgreSQL (for managing local storage) and Go
1.23+ for building the binary.

### Installation
The Github repo for this project is itself the package source of the
application. That is, to install, run

`go install github.com/BrandonIrizarry/gator@latest`

Then create a PostgreSQL database.

Finally, create a file named `.gatorconfig.json` with the following
content, where $CONN is the connection string for the database you
just created:

```
{
  "db_url": $CONN
}
```

## Usage

`./gator COMMAND ARGS`

## Commands

- `addfeed FEED-NAME FEED-URL`

    Add a feed to the local library of feeds, so that a user can later
    follow the feed if they choose.

    Right now, adding a feed automatically makes the currently logged-in
    user follow that feed.

- `agg FETCHING-INTERVAL`

    For each feed followed by the current user, fetch all of its posts
    into the local database, such that they'll be browseable later
    with `browse`.
        The idea is to leave this running as a background daemon, which
    would then fetch posts at some kind of reasonable interval (for
    example, once a week.)

- `browse [NUM-POSTS]`

    Output NUM-POSTS number of locally-saved posts in a pretty-printed
    format. The default value of NUM-POSTS is 2.

- `feeds`

    List all feeds by name, along with the user who added that feed.

- `follow FEED-URL`

    Make the currently logged-in user follow the indicated feed, such
    that the `agg` command (which see) will fetch posts from this
    feed.

- `following`

     Print out the list of feeds currently followed by the logged-in
     user.

- `login USERNAME`

    Set the currently logged-in user to USERNAME.

- `register USERNAME`

    Register USERNAME as a Gator user.

- `reset`

    Wipe all locally-saved RSS data clean (this command was mostly
    used in development for testing the database.)

- `users`

    List all registered users. The currently logged-in user is also
    specially indicated.

- `unfollow FEED-URL`

    Remove the feed (given by FEED-URL) from the current user's list
    of followed feeds, such that a subsequent `agg` operation won't
    fetch any more new feeds from there.
