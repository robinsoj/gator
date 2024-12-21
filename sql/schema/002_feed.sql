-- +goose Up
CREATE TABLE feeds (
	id UUID primary key,
	created_at timestamp not null,
	updated_at timestamp not null,
	name text not null unique,
	url text not null unique,
	user_id UUID not null,
	foreign key (user_id) references users(id) on delete cascade
);

-- +goose Down
DROP TABLE feeds;