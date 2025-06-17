-- name: CreateTransfer :one
INSERT INTO transfers (from_account_id, to_account_id, amount)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetTransfer :one
SELECT * FROM transfers
WHERE id = $1;

-- name: ListTransfers :many
SELECT *
FROM transfers
ORDER BY id
LIMIT $1 OFFSET $2;

-- name: ListTransfersByAccount :many
SELECT * FROM transfers
WHERE from_account_id = $1 OR to_account_id = $1
ORDER BY created_at DESC;

-- name: ListTransfersBetweenAccounts :many
SELECT * FROM transfers
WHERE from_account_id = $1 AND to_account_id = $2
ORDER BY created_at DESC;

-- name: UpdateTransferAmount :exec
UPDATE transfers
SET amount = $2
WHERE id = $1;

-- name: DeleteTransfer :exec
DELETE FROM transfers
WHERE id = $1;
