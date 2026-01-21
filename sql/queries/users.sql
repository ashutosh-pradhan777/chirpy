-- name: CreateUser :one
Insert into users (id, created_at, updated_at, email) 
Values (

    gen_random_uuid(),
    NOW(),
    NOW(),
    $1
)
Returning *;

-- name: ResetUser :exec
Delete from users;

