-- +goose Up
CREATE TABLE users (
	id UUID primary key,
	created_at timestamp not null,
	updated_at timestamp not null,
	name text not null unique
);

-- +goose Down
DROP TABLE users;
