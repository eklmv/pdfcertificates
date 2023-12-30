-- name: CreateStudent :one
INSERT INTO student (data)
VALUES ($1)
RETURNING *;

-- name: GetStudent :one
SELECT * FROM student
WHERE student_id = $1
LIMIT 1;

-- name: ListStudents :many
SELECT * FROM student
ORDER BY student_id
LIMIT $1 OFFSET $2;

-- name: UpdateStudent :one
UPDATE student
SET data = $2
WHERE student_id = $1
RETURNING *;

-- name: DeleteStudent :one
DELETE FROM student
WHERE student_id = $1
RETURNING *;
