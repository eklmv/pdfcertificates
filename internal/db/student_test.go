//go:build integration

package db

import (
	"context"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateStudent(t *testing.T) {
	t.Parallel()
	db := migrateUp(t)

	t.Run("nil data accepted, set as '{}", func(t *testing.T) {
		got, err := New().CreateStudent(context.Background(), db, nil)

		require.NoError(t, err)
		require.NotEmpty(t, got)
		assert.Equal(t, "{}", string(got.Data))
	})
	t.Run("non json string not accepted as data", func(t *testing.T) {
		got, err := New().CreateStudent(context.Background(), db, []byte("some string"))

		require.Error(t, err)
		require.Empty(t, got)
	})
	t.Run("invalid json not accepted as data", func(t *testing.T) {
		json := "{ 'key': 'value', }"
		got, err := New().CreateStudent(context.Background(), db, []byte(json))

		require.Error(t, err)
		require.Empty(t, got)
	})
	t.Run("valid json accepted as data", func(t *testing.T) {
		exp := randomData(t)

		got, err := New().CreateStudent(context.Background(), db, exp)

		require.NoError(t, err)
		require.NotEmpty(t, got)
		assert.Equal(t, exp, got.Data)
	})
}

func TestDeleteStudent(t *testing.T) {
	t.Parallel()
	db := migrateUp(t)

	exp, err := New().CreateStudent(context.Background(), db, randomData(t))
	require.NoError(t, err)
	require.NotEmpty(t, exp)

	got, err := New().DeleteStudent(context.Background(), db, exp.StudentID)

	require.NoError(t, err)
	require.NotEmpty(t, got)
	assert.Equal(t, exp, got)

	del, err := New().GetStudent(context.Background(), db, exp.StudentID)
	assert.Error(t, err)
	assert.Empty(t, del)
}

func TestGetStudent(t *testing.T) {
	t.Parallel()
	db := migrateUp(t)

	exp, err := New().CreateStudent(context.Background(), db, randomData(t))
	require.NoError(t, err)
	require.NotEmpty(t, exp)

	got, err := New().GetStudent(context.Background(), db, exp.StudentID)

	require.NoError(t, err)
	require.NotEmpty(t, got)
	require.Equal(t, exp, got)
}

func TestListStudents(t *testing.T) {
	t.Parallel()
	db := migrateUp(t)

	maxAmount := 100
	amount := rand.Intn(maxAmount) + 1
	exp := make([]Student, amount)
	for i := 0; i < amount; i++ {
		course, err := New().CreateStudent(context.Background(), db, randomData(t))
		require.NoError(t, err)
		require.NotEmpty(t, course)
		exp[i] = course
	}

	got, err := New().ListStudents(context.Background(), db, ListStudentsParams{
		Limit:  int64(amount),
		Offset: 0,
	})

	require.NoError(t, err)
	assert.NotEmpty(t, got)
	assert.ElementsMatch(t, exp, got)
}

func TestUpdateStudent(t *testing.T) {
	t.Parallel()
	db := migrateUp(t)

	t.Run("nil data accepted, set as '{}'", func(t *testing.T) {
		exp, err := New().CreateStudent(context.Background(), db, nil)
		require.NoError(t, err)
		require.NotEmpty(t, exp)
		newData := []byte("{}")
		exp.Data = newData

		got, err := New().UpdateStudent(context.Background(), db, UpdateStudentParams{
			StudentID: exp.StudentID,
			Data:      newData,
		})

		require.NoError(t, err)
		require.NotEmpty(t, got)
		assert.Equal(t, exp, got)
	})
	t.Run("valid json accepted", func(t *testing.T) {
		exp, err := New().CreateStudent(context.Background(), db, randomData(t))
		require.NoError(t, err)
		require.NotEmpty(t, exp)
		newData := randomData(t)
		exp.Data = newData

		got, err := New().UpdateStudent(context.Background(), db, UpdateStudentParams{
			StudentID: exp.StudentID,
			Data:      newData,
		})

		require.NoError(t, err)
		require.NotEmpty(t, got)
		assert.Equal(t, exp, got)
	})
}
