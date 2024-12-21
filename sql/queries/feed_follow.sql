-- name: CreateFeedFollow :one
with inserted_feed_follow as (
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
select insff.*, f.name as feed_name, u.name as user_name
from inserted_feed_follow insff inner join feeds f
on insff.feed_id = f.id
inner join users u
on insff.user_id = u.id;

-- name: GetFeedFollowsForUser :many
select f.name as feedname, u.name
from users u inner join feed_follows ff
on u.id = ff.user_id
inner join feeds f
on ff.feed_id = f.id
where u.name = $1;

-- name: DeleteFeedFollowForUser :exec
delete from feed_follows ff
using feeds f
where ff.user_id = $1 and ff.feed_id = f.id and f.url = $2;