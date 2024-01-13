-- name: CreateTemplate :one
INSERT INTO template (content)
VALUES ($1)
RETURNING *;

-- name: GetTemplate :one
SELECT * FROM template
WHERE template_id = $1
LIMIT 1;

-- name: ListTemplates :many
SELECT * FROM template
ORDER BY template_id
LIMIT $1 OFFSET $2;

-- name: ListTemplatesLen :one
SELECT count(*) FROM template;

-- name: UpdateTemplate :one
UPDATE template
SET content = $2
WHERE template_id = $1
RETURNING *;

-- name: DeleteTemplate :one
DELETE FROM template
WHERE template_id = $1
RETURNING *;
