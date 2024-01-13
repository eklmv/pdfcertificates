//go:build integration

package db

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func prepareCreateCertificateParams(tb testing.TB, db DBTX) CreateCertificateParams {
	tb.Helper()
	t, err := New().CreateTemplate(context.Background(), db, randomContent(tb))
	require.NoError(tb, err)
	require.NotEmpty(tb, t)
	c, err := New().CreateCourse(context.Background(), db, randomData(tb))
	require.NoError(tb, err)
	require.NotEmpty(tb, c)
	s, err := New().CreateStudent(context.Background(), db, randomData(tb))
	require.NoError(tb, err)
	require.NotEmpty(tb, s)
	return CreateCertificateParams{
		TemplateID: t.TemplateID,
		CourseID:   c.CourseID,
		StudentID:  s.StudentID,
	}
}

func randomCertificate(tb testing.TB, db DBTX) Certificate {
	tb.Helper()
	p := prepareCreateCertificateParams(tb, db)
	p.Data = randomData(tb)
	c, err := New().CreateCertificate(context.Background(), db, p)
	require.NoError(tb, err)
	require.NotEmpty(tb, c)
	return c
}

func TestCreateCertificate(t *testing.T) {
	t.Parallel()
	db := migrateUp(t)

	t.Run("nil data accepted, set as '{}", func(t *testing.T) {
		p := prepareCreateCertificateParams(t, db)
		p.Data = nil

		got, err := New().CreateCertificate(context.Background(), db, p)

		require.NoError(t, err)
		require.NotEmpty(t, got)
		assert.Equal(t, "{}", string(got.Data))
	})
	t.Run("non json string not accepted as data", func(t *testing.T) {
		p := prepareCreateCertificateParams(t, db)
		p.Data = []byte("some string")

		got, err := New().CreateCertificate(context.Background(), db, p)

		require.Error(t, err)
		require.Empty(t, got)
	})
	t.Run("invalid json not accepted as data", func(t *testing.T) {
		p := prepareCreateCertificateParams(t, db)
		p.Data = []byte("{ 'key': 'value', }")
		got, err := New().CreateCertificate(context.Background(), db, p)

		require.Error(t, err)
		require.Empty(t, got)
	})
	t.Run("valid json accepted as data", func(t *testing.T) {
		p := prepareCreateCertificateParams(t, db)
		p.Data = randomData(t)

		got, err := New().CreateCertificate(context.Background(), db, p)

		require.NoError(t, err)
		require.NotEmpty(t, got)
		assert.Equal(t, p.Data, got.Data)
	})
}

func TestGeneratedCertificateID(t *testing.T) {
	t.Parallel()
	db := migrateUp(t)

	got := randomCertificate(t, db)

	assert.Regexp(t, "^[0-9a-fA-F]{8}$", got.CertificateID)
}

func TestDeleteCertificate(t *testing.T) {
	t.Parallel()
	db := migrateUp(t)

	exp := randomCertificate(t, db)

	got, err := New().DeleteCertificate(context.Background(), db, exp.CertificateID)

	require.NoError(t, err)
	require.NotEmpty(t, got)
	assert.Equal(t, exp, got)

	del, err := New().GetCertificate(context.Background(), db, exp.CertificateID)
	assert.Error(t, err)
	assert.Empty(t, del)
}

func TestGetCertificate(t *testing.T) {
	t.Parallel()
	db := migrateUp(t)

	exp := randomCertificate(t, db)

	got, err := New().GetCertificate(context.Background(), db, exp.CertificateID)

	require.NoError(t, err)
	require.NotEmpty(t, got)
	require.Equal(t, exp, got)
}

func TestListCertificates(t *testing.T) {
	t.Parallel()
	db := migrateUp(t)

	maxAmount := 2
	amount := rand.Intn(maxAmount) + 1
	exp := make([]Certificate, amount)
	for i := 0; i < amount; i++ {
		c := randomCertificate(t, db)
		exp[i] = c
	}

	got, err := New().ListCertificates(context.Background(), db, ListCertificatesParams{
		Limit:  int64(amount),
		Offset: 0,
	})

	require.NoError(t, err)
	assert.NotEmpty(t, got)
	assert.ElementsMatch(t, exp, got)
}

func TestListCertificatesLen(t *testing.T) {
	t.Parallel()
	db := migrateUp(t)

	maxAmount := 10
	exp := rand.Intn(maxAmount) + 1
	for i := 0; i < exp; i++ {
		_ = randomCertificate(t, db)
	}

	got, err := New().ListCertificatesLen(context.Background(), db)

	require.NoError(t, err)
	assert.Equal(t, int64(exp), got)
}

func TestListCertificatesByCourse(t *testing.T) {
	t.Parallel()
	db := migrateUp(t)

	maxAmount := 100
	amount := rand.Intn(maxAmount) + 1
	exp := make([]Certificate, amount)
	expP := prepareCreateCertificateParams(t, db)
	for i := 0; i < amount; i++ {
		expP.Data = randomData(t)
		c, err := New().CreateCertificate(context.Background(), db, expP)
		require.NoError(t, err)
		require.NotEmpty(t, c)
		exp[i] = c
		randomCertificate(t, db)
	}
	got, err := New().ListCertificatesByCourse(context.Background(), db, ListCertificatesByCourseParams{
		CourseID: expP.CourseID,
		Limit:    int64(amount),
		Offset:   0})

	require.NoError(t, err)
	assert.NotEmpty(t, got)
	assert.ElementsMatch(t, exp, got)
}

func TestListCertificatesByCourseLen(t *testing.T) {
	t.Parallel()
	db := migrateUp(t)

	maxAmount := 10
	exp := rand.Intn(maxAmount) + 1
	p := prepareCreateCertificateParams(t, db)
	for i := 0; i < exp; i++ {
		p.Data = randomData(t)
		c, err := New().CreateCertificate(context.Background(), db, p)
		require.NoError(t, err)
		require.NotEmpty(t, c)
		randomCertificate(t, db)
	}
	got, err := New().ListCertificatesByCourseLen(context.Background(), db, p.CourseID)

	require.NoError(t, err)
	assert.Equal(t, int64(exp), got)
}

func TestListCertificatesByStudent(t *testing.T) {
	t.Parallel()
	db := migrateUp(t)

	maxAmount := 100
	amount := rand.Intn(maxAmount) + 1
	exp := make([]Certificate, amount)
	expP := prepareCreateCertificateParams(t, db)
	for i := 0; i < amount; i++ {
		expP.Data = randomData(t)
		c, err := New().CreateCertificate(context.Background(), db, expP)
		require.NoError(t, err)
		require.NotEmpty(t, c)
		exp[i] = c
		randomCertificate(t, db)
	}
	got, err := New().ListCertificatesByStudent(context.Background(), db, ListCertificatesByStudentParams{
		StudentID: expP.StudentID,
		Limit:     int64(amount),
		Offset:    0})

	require.NoError(t, err)
	assert.NotEmpty(t, got)
	assert.ElementsMatch(t, exp, got)
}

func TestListCertificatesByStudentLen(t *testing.T) {
	t.Parallel()
	db := migrateUp(t)

	maxAmount := 10
	exp := rand.Intn(maxAmount) + 1
	p := prepareCreateCertificateParams(t, db)
	for i := 0; i < exp; i++ {
		p.Data = randomData(t)
		c, err := New().CreateCertificate(context.Background(), db, p)
		require.NoError(t, err)
		require.NotEmpty(t, c)
		randomCertificate(t, db)
	}
	got, err := New().ListCertificatesByStudentLen(context.Background(), db, p.StudentID)

	require.NoError(t, err)
	assert.Equal(t, int64(exp), got)
}

func TestListCertificatesByTemplate(t *testing.T) {
	t.Parallel()
	db := migrateUp(t)

	maxAmount := 100
	amount := rand.Intn(maxAmount) + 1
	exp := make([]Certificate, amount)
	expP := prepareCreateCertificateParams(t, db)
	for i := 0; i < amount; i++ {
		expP.Data = randomData(t)
		c, err := New().CreateCertificate(context.Background(), db, expP)
		require.NoError(t, err)
		require.NotEmpty(t, c)
		exp[i] = c
		randomCertificate(t, db)
	}
	got, err := New().ListCertificatesByTemplate(context.Background(), db, ListCertificatesByTemplateParams{
		TemplateID: expP.TemplateID,
		Limit:      int64(amount),
		Offset:     0})

	require.NoError(t, err)
	assert.NotEmpty(t, got)
	assert.ElementsMatch(t, exp, got)
}

func TestListCertificatesByTemplateLen(t *testing.T) {
	t.Parallel()
	db := migrateUp(t)

	maxAmount := 10
	exp := rand.Intn(maxAmount) + 1
	p := prepareCreateCertificateParams(t, db)
	for i := 0; i < exp; i++ {
		p.Data = randomData(t)
		c, err := New().CreateCertificate(context.Background(), db, p)
		require.NoError(t, err)
		require.NotEmpty(t, c)
		randomCertificate(t, db)
	}
	got, err := New().ListCertificatesByTemplateLen(context.Background(), db, p.TemplateID)

	require.NoError(t, err)
	assert.Equal(t, int64(exp), got)
}

func TestUpdateCertificate(t *testing.T) {
	t.Parallel()
	db := migrateUp(t)

	t.Run("nil data accepted, set as '{}'", func(t *testing.T) {
		exp := randomCertificate(t, db)
		newData := []byte("{}")

		got, err := New().UpdateCertificate(context.Background(), db, UpdateCertificateParams{
			CertificateID: exp.CertificateID,
			Data:          newData,
		})

		require.NoError(t, err)
		require.NotEmpty(t, got)
		assert.Equal(t, exp.CertificateID, got.CertificateID)
		assert.Equal(t, exp.TemplateID, got.TemplateID)
		assert.Equal(t, exp.CourseID, got.CourseID)
		assert.Equal(t, exp.StudentID, got.StudentID)
		assert.Equal(t, newData, got.Data)
		assert.NotEqual(t, exp.Timestamp, got.Timestamp)
		assert.WithinDuration(t, exp.Timestamp.Time, got.Timestamp.Time, 1*time.Second)
	})
	t.Run("valid json accepted", func(t *testing.T) {
		exp := randomCertificate(t, db)
		newData := randomData(t)

		got, err := New().UpdateCertificate(context.Background(), db, UpdateCertificateParams{
			CertificateID: exp.CertificateID,
			Data:          newData,
		})

		require.NoError(t, err)
		require.NotEmpty(t, got)
		assert.Equal(t, exp.CertificateID, got.CertificateID)
		assert.Equal(t, exp.TemplateID, got.TemplateID)
		assert.Equal(t, exp.CourseID, got.CourseID)
		assert.Equal(t, exp.StudentID, got.StudentID)
		assert.Equal(t, newData, got.Data)
		assert.NotEqual(t, exp.Timestamp, got.Timestamp)
		assert.WithinDuration(t, exp.Timestamp.Time, got.Timestamp.Time, 1*time.Second)
	})
}

func TestUpdateCertificateTimestamp(t *testing.T) {
	t.Parallel()
	db := migrateUp(t)

	t.Run("updated certificate timestamp after updating template", func(t *testing.T) {
		cert := randomCertificate(t, db)

		_, err := New().UpdateTemplate(context.Background(), db, UpdateTemplateParams{
			TemplateID: cert.TemplateID,
			Content:    randomContent(t),
		})
		require.NoError(t, err)

		got, err := New().GetCertificate(context.Background(), db, cert.CertificateID)

		require.NoError(t, err)
		require.NotEmpty(t, got)
		assert.NotEqual(t, cert.Timestamp, got.Timestamp)
		assert.WithinDuration(t, cert.Timestamp.Time, got.Timestamp.Time, 1*time.Second)

	})
	t.Run("updated certificate timestamp after updating course", func(t *testing.T) {
		cert := randomCertificate(t, db)

		_, err := New().UpdateCourse(context.Background(), db, UpdateCourseParams{
			CourseID: cert.CourseID,
			Data:     randomData(t),
		})
		require.NoError(t, err)

		got, err := New().GetCertificate(context.Background(), db, cert.CertificateID)

		require.NoError(t, err)
		require.NotEmpty(t, got)
		assert.NotEqual(t, cert.Timestamp, got.Timestamp)
		assert.WithinDuration(t, cert.Timestamp.Time, got.Timestamp.Time, 1*time.Second)

	})
	t.Run("updated certificate timestamp after updating student", func(t *testing.T) {
		cert := randomCertificate(t, db)

		_, err := New().UpdateStudent(context.Background(), db, UpdateStudentParams{
			StudentID: cert.StudentID,
			Data:      randomData(t),
		})
		require.NoError(t, err)

		got, err := New().GetCertificate(context.Background(), db, cert.CertificateID)

		require.NoError(t, err)
		require.NotEmpty(t, got)
		assert.NotEqual(t, cert.Timestamp, got.Timestamp)
		assert.WithinDuration(t, cert.Timestamp.Time, got.Timestamp.Time, 1*time.Second)

	})
}
