-- name: CreateUser :one
Insert into users (id, created_at, updated_at, email, hashed_password) 
Values (

    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
)
Returning id, created_at, updated_at, email;

-- name: ResetUser :exec
Delete from users;

-- name: ReturnUser :one
Select * from users
where email = $1
LIMIT 1;