-- name: CreatePost :one
INSERT INTO posts (
    id, created_at, updated_at, title, url, description, published_at, feed_id
)
VALUES ($1, now(), now(), $2, $3, $4, $5, $6)
ON CONFLICT (url) DO NOTHING
RETURNING *;

-- name: GetPostsByUser :many
SELECT posts.title, posts.url, posts.description, posts.published_at
FROM posts
INNER JOIN feeds
    ON feeds.id = posts.feed_id
INNER JOIN users
    ON feeds.user_id = users.id
INNER JOIN feed_follows
    ON feed_follows.feed_id = posts.feed_id
WHERE feed_follows.user_id = $1
ORDER BY published_at DESC NULLS LAST
LIMIT $2;
