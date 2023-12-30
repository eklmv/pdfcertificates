-- name: CreateCertificate :one
INSERT INTO certificate (template_id, course_id, student_id, data)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetCertificate :one
SELECT * FROM certificate
WHERE certificate_id = $1
LIMIT 1;

-- name: ListCertificates :many
SELECT * FROM certificate
ORDER BY certificate_id
LIMIT $1 OFFSET $2;

-- name: ListCertificatesByTemplate :many
SELECT * FROM certificate
WHERE template_id = $1
ORDER BY certificate_id
LIMIT $2 OFFSET $3;

-- name: ListCertificatesByCourse :many
SELECT * FROM certificate
WHERE course_id = $1
ORDER BY certificate_id
LIMIT $2 OFFSET $3;

-- name: ListCertificatesByStudent :many
SELECT * FROM certificate
WHERE student_id = $1
ORDER BY certificate_id
LIMIT $2 OFFSET $3;

-- name: UpdateCertificate :one
UPDATE certificate
SET data = $2
WHERE certificate_id = $1
RETURNING *;

-- name: DeleteCertificate :one
DELETE FROM certificate
WHERE certificate_id = $1
RETURNING *;
