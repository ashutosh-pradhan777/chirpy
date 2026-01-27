-- name: CreateChirp :one
Insert into chirps (id, created_at, updated_at, body, user_id) 
Values (

    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
)
Returning *;

-- name: ReturnAllChirps :many
Select * from chirps
order by created_at ASC;

-- name: ReturnChirp :one
Select * from chirps
where id = $1
LIMIT 1;


