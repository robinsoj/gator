-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, name)
VALUES (
    $1,
    $2,
    $3,
    $4
)
RETURNING *;

-- name: GetUser :one
select id, created_at, updated_at, name
from users
where $1 = name
limit 1;

-- name: DeleteUsers :exec
delete from users;

-- name: GetUsers :many
select name
from users;