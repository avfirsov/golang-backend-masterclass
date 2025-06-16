-- name: CreateAccount :one
INSERT INTO accounts (
    balance, owner, currency
) VALUES (
             $1, $2, $3
         )
RETURNING *;