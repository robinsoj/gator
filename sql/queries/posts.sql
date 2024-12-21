-- name: CreatePost :exec
insert into posts (id, created_at, updated_at, title, url, description, published_at, feed_id)
values (
	$1,
	$2,
	$3,
	$4,
	$5,
	$6,
	$7,
	$8
);

-- name: GetPostsForUser :many
select p.*
from posts p inner join feed_follows ff
on p.feed_id = ff.feed_id
inner join users u
on ff.user_id = u.id
where u.name = $1
order by p.published_at desc
limit $2;