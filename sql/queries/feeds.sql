-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
)
RETURNING *;

-- name: ListFeeds :many
select f.name, f.url, u.name
from feeds f, users u
where f.user_id = u.id;

-- name: ListFeedsByURL :many
select *
from feeds
where url = $1;

-- name: MarkFeedFetched :exec
update feeds
set updated_at = now(), last_fetched_at = now()
where id = $1;

-- name: GetNextFeedToFetch :one
select *
from feeds
order by last_fetched_at nulls first
limit 1;