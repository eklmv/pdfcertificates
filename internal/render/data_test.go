package render

import (
	"testing"
	"time"

	"github.com/eklmv/pdfcertificates/internal/db"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractData(t *testing.T) {
	host := "https://localhost/cert/"
	expKey := "key"
	expValue := "test"
	data := []byte("{ \"" + expKey + "\": \"" + expValue + "\" }")
	course := db.Course{
		CourseID: 0,
		Data:     data,
	}
	student := db.Student{
		StudentID: 0,
		Data:      data,
	}
	cert := db.Certificate{
		CertificateID: "00000000",
		TemplateID:    0,
		CourseID:      course.CourseID,
		StudentID:     student.StudentID,
		Timestamp:     pgtype.Timestamptz{Time: time.Now(), Valid: true},
		Data:          data,
	}

	got, err := ExtractData(host, cert, course, student)

	require.NoError(t, err)
	require.NotEmpty(t, got)
	assert.Equal(t, cert.CertificateID, got.CertificateID)
	assert.Equal(t, host+cert.CertificateID, got.Link)
	require.Contains(t, got.Certificate, expKey)
	require.Contains(t, got.Course, expKey)
	require.Contains(t, got.Student, expKey)
	assert.Equal(t, expValue, got.Certificate[expKey])
	assert.Equal(t, expValue, got.Course[expKey])
	assert.Equal(t, expValue, got.Student[expKey])
}
