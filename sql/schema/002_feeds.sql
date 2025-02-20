-- +goose Up
CREATE TABLE feeds(
       id UUID PRIMARY KEY,
       created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
       updated_at TIMESTAMP NOT NULL,
       name TEXT NOT NULL, -- the name/title of the given RSS feed
       url TEXT UNIQUE NOT NULL,
       user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE
);


-- +goose Down
DROP TABLE feeds;
