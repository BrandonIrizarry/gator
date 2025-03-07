-- name: CreateFeedFollow :one
WITH inserted_feed_follow AS (
     INSERT INTO feed_follows (id, created_at, updated_at, user_id, feed_id)
     VALUES (
            $1,
            $2,
            $3,
            $4,
            $5
     )

     RETURNING *
)
SELECT inserted_feed_follow.*, feeds.name AS feedname, users.name AS username
FROM inserted_feed_follow
INNER JOIN feeds
ON feeds.id = inserted_feed_follow.feed_id
INNER JOIN users
ON users.id = inserted_feed_follow.user_id;

-- name: GetFeedFollowsForUser :many
SELECT feed_follows.*, feeds.name AS feedname
FROM feed_follows
INNER JOIN feeds
ON feeds.id = feed_follows.feed_id
INNER JOIN users
ON users.id = feed_follows.user_id
WHERE users.id = $1;

-- name: DeleteFeedFollow :execrows
DELETE FROM feed_follows USING feeds
WHERE feed_follows.user_id = $1 AND feeds.url = $2;

-- name: GetNextFeedToFetch :many
SELECT * FROM feed_follows
INNER JOIN feeds
ON feeds.id = feed_follows.feed_id
ORDER BY feeds.last_fetched_at NULLS FIRST;

