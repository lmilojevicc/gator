-- name: CreateFeed :exec
INSERT INTO feeds (id, name, url, user_id)
VALUES ($1, $2, $3, $4);

-- name: GetAllFeeds :many
SELECT * FROM feeds;
