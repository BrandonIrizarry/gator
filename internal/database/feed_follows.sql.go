// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: feed_follows.sql

package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

const createFeedFollow = `-- name: CreateFeedFollow :one
WITH inserted_feed_follow AS (
     INSERT INTO feed_follows (id, created_at, updated_at, user_id, feed_id)
     VALUES (
            $1,
            $2,
            $3,
            $4,
            $5
     )

     RETURNING id, created_at, updated_at, user_id, feed_id
)
SELECT inserted_feed_follow.id, inserted_feed_follow.created_at, inserted_feed_follow.updated_at, inserted_feed_follow.user_id, inserted_feed_follow.feed_id, feeds.name AS feedname, users.name AS username
FROM inserted_feed_follow
INNER JOIN feeds
ON feeds.id = inserted_feed_follow.feed_id
INNER JOIN users
ON users.id = inserted_feed_follow.user_id
`

type CreateFeedFollowParams struct {
	ID        uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	UserID    uuid.UUID
	FeedID    uuid.UUID
}

type CreateFeedFollowRow struct {
	ID        uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	UserID    uuid.UUID
	FeedID    uuid.UUID
	Feedname  string
	Username  string
}

func (q *Queries) CreateFeedFollow(ctx context.Context, arg CreateFeedFollowParams) (CreateFeedFollowRow, error) {
	row := q.db.QueryRowContext(ctx, createFeedFollow,
		arg.ID,
		arg.CreatedAt,
		arg.UpdatedAt,
		arg.UserID,
		arg.FeedID,
	)
	var i CreateFeedFollowRow
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.UserID,
		&i.FeedID,
		&i.Feedname,
		&i.Username,
	)
	return i, err
}

const deleteFeedFollow = `-- name: DeleteFeedFollow :execrows
DELETE FROM feed_follows USING feeds
WHERE feed_follows.user_id = $1 AND feeds.url = $2
`

type DeleteFeedFollowParams struct {
	UserID uuid.UUID
	Url    string
}

func (q *Queries) DeleteFeedFollow(ctx context.Context, arg DeleteFeedFollowParams) (int64, error) {
	result, err := q.db.ExecContext(ctx, deleteFeedFollow, arg.UserID, arg.Url)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const getFeedFollowsForUser = `-- name: GetFeedFollowsForUser :many
SELECT feed_follows.id, feed_follows.created_at, feed_follows.updated_at, feed_follows.user_id, feed_follows.feed_id, feeds.name AS feedname
FROM feed_follows
INNER JOIN feeds
ON feeds.id = feed_follows.feed_id
INNER JOIN users
ON users.id = feed_follows.user_id
WHERE users.id = $1
`

type GetFeedFollowsForUserRow struct {
	ID        uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	UserID    uuid.UUID
	FeedID    uuid.UUID
	Feedname  string
}

func (q *Queries) GetFeedFollowsForUser(ctx context.Context, id uuid.UUID) ([]GetFeedFollowsForUserRow, error) {
	rows, err := q.db.QueryContext(ctx, getFeedFollowsForUser, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetFeedFollowsForUserRow
	for rows.Next() {
		var i GetFeedFollowsForUserRow
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.UserID,
			&i.FeedID,
			&i.Feedname,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getNextFeedToFetch = `-- name: GetNextFeedToFetch :many
SELECT feed_follows.id, feed_follows.created_at, feed_follows.updated_at, feed_follows.user_id, feed_id, feeds.id, feeds.created_at, feeds.updated_at, name, url, feeds.user_id, last_fetched_at FROM feed_follows
INNER JOIN feeds
ON feeds.id = feed_follows.feed_id
ORDER BY feeds.last_fetched_at NULLS FIRST
`

type GetNextFeedToFetchRow struct {
	ID            uuid.UUID
	CreatedAt     time.Time
	UpdatedAt     time.Time
	UserID        uuid.UUID
	FeedID        uuid.UUID
	ID_2          uuid.UUID
	CreatedAt_2   time.Time
	UpdatedAt_2   time.Time
	Name          string
	Url           string
	UserID_2      uuid.UUID
	LastFetchedAt sql.NullTime
}

func (q *Queries) GetNextFeedToFetch(ctx context.Context) ([]GetNextFeedToFetchRow, error) {
	rows, err := q.db.QueryContext(ctx, getNextFeedToFetch)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetNextFeedToFetchRow
	for rows.Next() {
		var i GetNextFeedToFetchRow
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.UserID,
			&i.FeedID,
			&i.ID_2,
			&i.CreatedAt_2,
			&i.UpdatedAt_2,
			&i.Name,
			&i.Url,
			&i.UserID_2,
			&i.LastFetchedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
