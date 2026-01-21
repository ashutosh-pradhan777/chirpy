-- +goose Up
Create Table users (
    id UUID primary key,
    created_at Timestamp not null,
    updated_at Timestamp not null,
    email Text not null unique
);

-- +goose Down
DROP TABLE users;

--