-- +goose Up
Create Table chirps (
    id UUID primary key,
    created_at Timestamp not null,
    updated_at Timestamp not null,
    body Text not null unique,
    user_id UUID References users(id) On Delete Cascade
);

-- +goose Down
DROP TABLE chirps;