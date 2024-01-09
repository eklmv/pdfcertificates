// go.build integration

package db

import (
	"context"
	"log/slog"
	"math/rand"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func randomContent(tb testing.TB) string {
	tb.Helper()
	content, err := gofakeit.XML(nil)
	if err != nil {
		slog.Error("random conten generation failed", slog.Any("err", err))
		tb.FailNow()
	}
	return string(content)
}

func TestCreateTemplate(t *testing.T) {
	t.Parallel()
	db := migrateUp(t)

	t.Run("content can not be empty string", func(t *testing.T) {

		got, err := New().CreateTemplate(context.Background(), db, "")

		assert.Error(t, err)
		assert.Empty(t, got)
	})
	t.Run("content accepts random non empty string", func(t *testing.T) {
		exp := randomContent(t)

		got, err := New().CreateTemplate(context.Background(), db, exp)

		require.NoError(t, err)
		require.NotEmpty(t, got)
		assert.NotEmpty(t, got.TemplateID)
		assert.Equal(t, exp, got.Content)
	})
}

func TestDeleteTemplate(t *testing.T) {
	t.Parallel()
	db := migrateUp(t)

	exp, err := New().CreateTemplate(context.Background(), db, randomContent(t))
	require.NoError(t, err)
	require.NotEmpty(t, exp)

	got, err := New().DeleteTemplate(context.Background(), db, exp.TemplateID)

	require.NoError(t, err)
	require.NotEmpty(t, got)
	assert.Equal(t, exp, got)

	del, err := New().GetTemplate(context.Background(), db, exp.TemplateID)
	assert.Error(t, err)
	assert.Empty(t, del)
}

func TestGetTemplate(t *testing.T) {
	t.Parallel()
	db := migrateUp(t)

	exp, err := New().CreateTemplate(context.Background(), db, randomContent(t))
	require.NoError(t, err)
	require.NotEmpty(t, exp)

	got, err := New().GetTemplate(context.Background(), db, exp.TemplateID)

	require.NoError(t, err)
	require.NotEmpty(t, got)
	require.Equal(t, exp, got)
}

func TestListTemplates(t *testing.T) {
	t.Parallel()
	db := migrateUp(t)

	maxAmount := 100
	amount := rand.Intn(maxAmount) + 1
	exp := make([]Template, amount)
	for i := 0; i < amount; i++ {
		tmpl, err := New().CreateTemplate(context.Background(), db, randomContent(t))
		require.NoError(t, err)
		require.NotEmpty(t, tmpl)
		exp[i] = tmpl
	}

	got, err := New().ListTemplates(context.Background(), db, ListTemplatesParams{
		Limit:  int64(amount),
		Offset: 0,
	})

	require.NoError(t, err)
	assert.NotEmpty(t, got)
	assert.ElementsMatch(t, exp, got)
}

func TestUpdateTemplate(t *testing.T) {
	t.Parallel()
	db := migrateUp(t)

	exp, err := New().CreateTemplate(context.Background(), db, randomContent(t))
	require.NoError(t, err)
	require.NotEmpty(t, exp)
	newContent := randomContent(t)
	exp.Content = newContent

	got, err := New().UpdateTemplate(context.Background(), db, UpdateTemplateParams{
		TemplateID: exp.TemplateID,
		Content:    newContent,
	})

	require.NoError(t, err)
	require.NotEmpty(t, got)
	assert.Equal(t, exp, got)
}
