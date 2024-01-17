package db

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"
	"unsafe"

	"github.com/eklmv/pdfcertificates/internal/cache"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func prepCachedQueries(tb testing.TB) (cq *CachedQueries, c cache.Cache[uint32, cachedResponse], m *MockQuerier) {
	tb.Helper()
	m = NewMockQuerier(tb)
	c = cache.NewSafeCache(cache.NewLRUCache[uint32, cachedResponse](0))
	cq = NewCachedQueries(c, m)
	return
}

func TestCachedQueriesImplementsInterface(t *testing.T) {
	assert.Implements(t, (*Querier)(nil), &CachedQueries{})
}

func TestCachedQueriesCreateCertificate(t *testing.T) {
	t.Run("if certificate successfully created it should be cached", func(t *testing.T) {
		cq, c, m := prepCachedQueries(t)
		ctx := context.Background()
		exp := Certificate{
			CertificateID: "00000000",
			TemplateID:    0,
			CourseID:      0,
			StudentID:     0,
			Timestamp: pgtype.Timestamptz{
				Time:             time.Now(),
				InfinityModifier: 0,
				Valid:            true,
			},
			Data: []byte{},
		}
		m.EXPECT().CreateCertificate(ctx, nil, CreateCertificateParams{}).Return(exp, nil).Once()

		got, err := cq.CreateCertificate(ctx, nil, CreateCertificateParams{})

		assert.NoError(t, err)
		assert.Equal(t, exp, got)
		m.AssertExpectations(t)

		hash, hErr := cache.HashString(prefCert.String() + exp.CertificateID)
		require.NoError(t, hErr)
		assert.Contains(t, c.Keys(), hash)
		assert.Equal(t, uint64(1), c.Len())
		assert.Equal(t, uint64(unsafe.Sizeof(exp))+uint64(len(exp.CertificateID)), c.Size())
	})
}

func TestCachedQueriesCreateCourse(t *testing.T) {
	t.Run("if course successfully created it should be cached", func(t *testing.T) {
		cq, c, m := prepCachedQueries(t)
		ctx := context.Background()
		exp := Course{
			CourseID: 0,
			Data:     []byte("{}"),
		}
		m.EXPECT().CreateCourse(ctx, nil, exp.Data).Return(exp, nil).Once()

		got, err := cq.CreateCourse(ctx, nil, exp.Data)

		assert.NoError(t, err)
		assert.Equal(t, exp, got)
		m.AssertExpectations(t)

		hash, hErr := cache.HashString(prefCourse.String() + strconv.Itoa(int(exp.CourseID)))
		require.NoError(t, hErr)
		assert.Contains(t, c.Keys(), hash)
		assert.Equal(t, uint64(1), c.Len())
		assert.Equal(t, uint64(unsafe.Sizeof(exp))+uint64(len(exp.Data)), c.Size())
	})
}

func TestCachedQueriesCreateStudent(t *testing.T) {
	t.Run("if student successfully created it should be cached", func(t *testing.T) {
		cq, c, m := prepCachedQueries(t)
		ctx := context.Background()
		exp := Student{
			StudentID: 0,
			Data:      []byte("{}"),
		}
		m.EXPECT().CreateStudent(ctx, nil, exp.Data).Return(exp, nil).Once()

		got, err := cq.CreateStudent(ctx, nil, exp.Data)

		assert.NoError(t, err)
		assert.Equal(t, exp, got)
		m.AssertExpectations(t)

		hash, hErr := cache.HashString(prefStudent.String() + strconv.Itoa(int(exp.StudentID)))
		require.NoError(t, hErr)
		assert.Contains(t, c.Keys(), hash)
		assert.Equal(t, uint64(1), c.Len())
		assert.Equal(t, uint64(unsafe.Sizeof(exp))+uint64(len(exp.Data)), c.Size())
	})
}

func TestCachedQueriesCreateTemplate(t *testing.T) {
	t.Run("if template successfully created it should be cached", func(t *testing.T) {
		cq, c, m := prepCachedQueries(t)
		ctx := context.Background()
		exp := Template{
			TemplateID: 0,
			Content:    "123",
		}
		m.EXPECT().CreateTemplate(ctx, nil, exp.Content).Return(exp, nil).Once()

		got, err := cq.CreateTemplate(ctx, nil, exp.Content)

		assert.NoError(t, err)
		assert.Equal(t, exp, got)
		m.AssertExpectations(t)

		hash, hErr := cache.HashString(prefTmpl.String() + strconv.Itoa(int(exp.TemplateID)))
		require.NoError(t, hErr)
		assert.Contains(t, c.Keys(), hash)
		assert.Equal(t, uint64(1), c.Len())
		assert.Equal(t, uint64(unsafe.Sizeof(exp))+uint64(len(exp.Content)), c.Size())
	})
}

func TestCachedQueriesDeleteCertificate(t *testing.T) {
	t.Run("if certificate successfully deleted it should be removed from cache", func(t *testing.T) {
		cq, c, m := prepCachedQueries(t)
		ctx := context.Background()
		exp := Certificate{
			CertificateID: "00000000",
			TemplateID:    0,
			CourseID:      0,
			StudentID:     0,
			Timestamp: pgtype.Timestamptz{
				Time:             time.Now(),
				InfinityModifier: 0,
				Valid:            true,
			},
			Data: []byte{},
		}
		m.EXPECT().CreateCertificate(ctx, nil, CreateCertificateParams{}).Return(exp, nil).Once()
		_, err := cq.CreateCertificate(ctx, nil, CreateCertificateParams{})
		require.NoError(t, err)

		m.EXPECT().DeleteCertificate(ctx, nil, exp.CertificateID).Return(exp, nil).Once()
		got, err := cq.DeleteCertificate(ctx, nil, exp.CertificateID)

		assert.NoError(t, err)
		assert.Equal(t, exp, got)
		assert.Zero(t, c.Len())
		assert.Zero(t, c.Size())
		m.AssertExpectations(t)
	})
	t.Run("if certificate deletion failed cache should be intact", func(t *testing.T) {
		cq, c, m := prepCachedQueries(t)
		ctx := context.Background()
		cert := Certificate{
			CertificateID: "00000000",
			TemplateID:    0,
			CourseID:      0,
			StudentID:     0,
			Timestamp: pgtype.Timestamptz{
				Time:             time.Now(),
				InfinityModifier: 0,
				Valid:            true,
			},
			Data: []byte{},
		}
		m.EXPECT().CreateCertificate(ctx, nil, CreateCertificateParams{}).Return(cert, nil).Once()
		_, err := cq.CreateCertificate(ctx, nil, CreateCertificateParams{})
		require.NoError(t, err)

		m.EXPECT().DeleteCertificate(ctx, nil, cert.CertificateID).Return(Certificate{}, fmt.Errorf("failed")).Once()
		got, err := cq.DeleteCertificate(ctx, nil, cert.CertificateID)

		assert.ErrorContains(t, err, "failed")
		assert.Empty(t, got)
		assert.Equal(t, uint64(1), c.Len())
		assert.Equal(t, uint64(unsafe.Sizeof(cert))+uint64(len(cert.CertificateID)), c.Size())
		m.AssertExpectations(t)
	})
}

func TestCachedQueriesDeleteCourse(t *testing.T) {
	t.Run("if course successfully deleted it should be removed from cache", func(t *testing.T) {
		cq, c, m := prepCachedQueries(t)
		ctx := context.Background()
		exp := Course{
			CourseID: 0,
			Data:     []byte("{}"),
		}
		m.EXPECT().CreateCourse(ctx, nil, exp.Data).Return(exp, nil).Once()
		_, err := cq.CreateCourse(ctx, nil, exp.Data)
		require.NoError(t, err)

		m.EXPECT().DeleteCourse(ctx, nil, exp.CourseID).Return(exp, nil).Once()
		got, err := cq.DeleteCourse(ctx, nil, exp.CourseID)

		assert.NoError(t, err)
		assert.Equal(t, exp, got)
		assert.Zero(t, c.Len())
		assert.Zero(t, c.Size())
		m.AssertExpectations(t)
	})
	t.Run("if course deletion failed cache should be intact", func(t *testing.T) {
		cq, c, m := prepCachedQueries(t)
		ctx := context.Background()
		course := Course{
			CourseID: 0,
			Data:     []byte("{}"),
		}
		m.EXPECT().CreateCourse(ctx, nil, course.Data).Return(course, nil).Once()
		_, err := cq.CreateCourse(ctx, nil, course.Data)
		require.NoError(t, err)

		m.EXPECT().DeleteCourse(ctx, nil, course.CourseID).Return(Course{}, fmt.Errorf("failed")).Once()
		got, err := cq.DeleteCourse(ctx, nil, course.CourseID)

		assert.ErrorContains(t, err, "failed")
		assert.Empty(t, got)
		assert.Equal(t, uint64(1), c.Len())
		assert.Equal(t, uint64(unsafe.Sizeof(course))+uint64(len(course.Data)), c.Size())
		m.AssertExpectations(t)
	})
}

func TestCachedQueriesDeleteStudent(t *testing.T) {
	t.Run("if student successfully deleted it and all linked certificates should be removed from cache", func(t *testing.T) {
		cq, c, m := prepCachedQueries(t)
		ctx := context.Background()
		exp := Student{
			StudentID: 0,
			Data:      []byte("{}"),
		}
		cert := Certificate{
			CertificateID: "00000000",
			TemplateID:    0,
			CourseID:      0,
			StudentID:     exp.StudentID,
			Timestamp: pgtype.Timestamptz{
				Time:             time.Now(),
				InfinityModifier: 0,
				Valid:            true,
			},
			Data: []byte{},
		}
		m.EXPECT().CreateStudent(ctx, nil, exp.Data).Return(exp, nil).Once()
		_, err := cq.CreateStudent(ctx, nil, exp.Data)
		require.NoError(t, err)

		m.EXPECT().CreateCertificate(ctx, nil, CreateCertificateParams{}).Return(cert, nil).Once()
		_, err = cq.CreateCertificate(ctx, nil, CreateCertificateParams{})
		require.NoError(t, err)
		require.Equal(t, uint64(2), c.Len())

		m.EXPECT().DeleteStudent(ctx, nil, exp.StudentID).Return(exp, nil).Once()
		m.EXPECT().ListCertificatesByStudentLen(ctx, nil, exp.StudentID).Return(1, nil).Once()
		m.EXPECT().ListCertificatesByStudent(ctx, nil, ListCertificatesByStudentParams{
			StudentID: exp.StudentID,
			Limit:     1,
			Offset:    0,
		}).Return([]Certificate{cert}, nil).Once()
		got, err := cq.DeleteStudent(ctx, nil, exp.StudentID)

		assert.NoError(t, err)
		assert.Equal(t, exp, got)
		assert.Zero(t, c.Len())
		assert.Zero(t, c.Size())
		m.AssertExpectations(t)
	})
	t.Run("if student deletion failed cache should be intact", func(t *testing.T) {
		cq, c, m := prepCachedQueries(t)
		ctx := context.Background()
		student := Student{
			StudentID: 0,
			Data:      []byte("{}"),
		}
		cert := Certificate{
			CertificateID: "00000000",
			TemplateID:    0,
			CourseID:      0,
			StudentID:     student.StudentID,
			Timestamp: pgtype.Timestamptz{
				Time:             time.Now(),
				InfinityModifier: 0,
				Valid:            true,
			},
			Data: []byte{},
		}
		m.EXPECT().CreateStudent(ctx, nil, student.Data).Return(student, nil).Once()
		_, err := cq.CreateStudent(ctx, nil, student.Data)
		require.NoError(t, err)

		m.EXPECT().CreateCertificate(ctx, nil, CreateCertificateParams{}).Return(cert, nil).Once()
		_, err = cq.CreateCertificate(ctx, nil, CreateCertificateParams{})
		require.NoError(t, err)
		require.Equal(t, uint64(2), c.Len())

		m.EXPECT().DeleteStudent(ctx, nil, student.StudentID).Return(Student{}, fmt.Errorf("failed")).Once()
		got, err := cq.DeleteStudent(ctx, nil, student.StudentID)

		assert.ErrorContains(t, err, "failed")
		assert.Empty(t, got)
		assert.Equal(t, uint64(2), c.Len())
		assert.Equal(t, uint64(unsafe.Sizeof(student)+unsafe.Sizeof(cert))+
			uint64(len(student.Data)+len(cert.CertificateID)), c.Size())
		m.AssertExpectations(t)
	})
}

func TestCachedQueriesDeleteTemplate(t *testing.T) {
	t.Run("if template successfully deleted it should be removed from cache", func(t *testing.T) {
		cq, c, m := prepCachedQueries(t)
		ctx := context.Background()
		exp := Template{
			TemplateID: 0,
			Content:    "",
		}
		m.EXPECT().CreateTemplate(ctx, nil, exp.Content).Return(exp, nil).Once()
		_, err := cq.CreateTemplate(ctx, nil, exp.Content)
		require.NoError(t, err)

		m.EXPECT().DeleteTemplate(ctx, nil, exp.TemplateID).Return(exp, nil).Once()
		got, err := cq.DeleteTemplate(ctx, nil, exp.TemplateID)

		assert.NoError(t, err)
		assert.Equal(t, exp, got)
		assert.Zero(t, c.Len())
		assert.Zero(t, c.Size())
		m.AssertExpectations(t)
	})
	t.Run("if template deletion failed cache should be intact", func(t *testing.T) {
		cq, c, m := prepCachedQueries(t)
		ctx := context.Background()
		tmpl := Template{
			TemplateID: 0,
			Content:    "",
		}
		m.EXPECT().CreateTemplate(ctx, nil, tmpl.Content).Return(tmpl, nil).Once()
		_, err := cq.CreateTemplate(ctx, nil, tmpl.Content)
		require.NoError(t, err)

		m.EXPECT().DeleteTemplate(ctx, nil, tmpl.TemplateID).Return(Template{}, fmt.Errorf("failed")).Once()
		got, err := cq.DeleteTemplate(ctx, nil, tmpl.TemplateID)

		assert.ErrorContains(t, err, "failed")
		assert.Empty(t, got)
		assert.Equal(t, uint64(1), c.Len())
		assert.Equal(t, uint64(unsafe.Sizeof(tmpl))+uint64(len(tmpl.Content)), c.Size())
		m.AssertExpectations(t)
	})
}

func TestCachedQueriesGetCertificate(t *testing.T) {
	t.Run("if certificate cached, return it from cache and avoid db request", func(t *testing.T) {
		cq, _, m := prepCachedQueries(t)
		ctx := context.Background()
		exp := Certificate{
			CertificateID: "00000000",
			TemplateID:    0,
			CourseID:      0,
			StudentID:     0,
			Timestamp: pgtype.Timestamptz{
				Time:             time.Now(),
				InfinityModifier: 0,
				Valid:            true,
			},
			Data: []byte{},
		}
		m.EXPECT().CreateCertificate(ctx, nil, CreateCertificateParams{}).Return(exp, nil).Once()
		_, err := cq.CreateCertificate(ctx, nil, CreateCertificateParams{})
		require.NoError(t, err)

		got, err := cq.GetCertificate(ctx, nil, exp.CertificateID)

		assert.NoError(t, err)
		assert.Equal(t, exp, got)
		m.AssertExpectations(t)
	})
	t.Run("if cache empty, make db request and cache result", func(t *testing.T) {
		cq, c, m := prepCachedQueries(t)
		ctx := context.Background()
		exp := Certificate{
			CertificateID: "00000000",
			TemplateID:    0,
			CourseID:      0,
			StudentID:     0,
			Timestamp: pgtype.Timestamptz{
				Time:             time.Now(),
				InfinityModifier: 0,
				Valid:            true,
			},
			Data: []byte{},
		}

		m.EXPECT().GetCertificate(ctx, nil, exp.CertificateID).Return(exp, nil).Once()
		got, err := cq.GetCertificate(ctx, nil, exp.CertificateID)

		assert.NoError(t, err)
		assert.Equal(t, exp, got)
		m.AssertExpectations(t)

		hash, hErr := cache.HashString(prefCert.String() + exp.CertificateID)
		require.NoError(t, hErr)
		assert.Contains(t, c.Keys(), hash)
		assert.Equal(t, uint64(1), c.Len())
		assert.Equal(t, uint64(unsafe.Sizeof(exp))+uint64(len(exp.CertificateID)), c.Size())
	})
}

func TestCachedQueriesGetCourse(t *testing.T) {
	t.Run("if course cached, return it from cache and avoid db request", func(t *testing.T) {
		cq, _, m := prepCachedQueries(t)
		ctx := context.Background()
		exp := Course{
			CourseID: 0,
			Data:     []byte("{}"),
		}
		m.EXPECT().CreateCourse(ctx, nil, exp.Data).Return(exp, nil).Once()
		_, err := cq.CreateCourse(ctx, nil, exp.Data)
		require.NoError(t, err)

		got, err := cq.GetCourse(ctx, nil, exp.CourseID)

		assert.NoError(t, err)
		assert.Equal(t, exp, got)
		m.AssertExpectations(t)
	})
	t.Run("if cache empty, make db request and cache result", func(t *testing.T) {
		cq, c, m := prepCachedQueries(t)
		ctx := context.Background()
		exp := Course{
			CourseID: 0,
			Data:     []byte("{}"),
		}

		m.EXPECT().GetCourse(ctx, nil, exp.CourseID).Return(exp, nil).Once()
		got, err := cq.GetCourse(ctx, nil, exp.CourseID)

		assert.NoError(t, err)
		assert.Equal(t, exp, got)
		m.AssertExpectations(t)

		hash, hErr := cache.HashString(prefCourse.String() + strconv.Itoa(int(exp.CourseID)))
		require.NoError(t, hErr)
		assert.Contains(t, c.Keys(), hash)
		assert.Equal(t, uint64(1), c.Len())
		assert.Equal(t, uint64(unsafe.Sizeof(exp))+uint64(len(exp.Data)), c.Size())
	})
}

func TestCachedQueriesGetStudent(t *testing.T) {
	t.Run("if student cached, return it from cache and avoid db request", func(t *testing.T) {
		cq, _, m := prepCachedQueries(t)
		ctx := context.Background()
		exp := Student{
			StudentID: 0,
			Data:      []byte("{}"),
		}
		m.EXPECT().CreateStudent(ctx, nil, exp.Data).Return(exp, nil).Once()
		_, err := cq.CreateStudent(ctx, nil, exp.Data)
		require.NoError(t, err)

		got, err := cq.GetStudent(ctx, nil, exp.StudentID)

		assert.NoError(t, err)
		assert.Equal(t, exp, got)
		m.AssertExpectations(t)
	})
	t.Run("if cache empty, make db request and cache result", func(t *testing.T) {
		cq, c, m := prepCachedQueries(t)
		ctx := context.Background()
		exp := Student{
			StudentID: 0,
			Data:      []byte("{}"),
		}

		m.EXPECT().GetStudent(ctx, nil, exp.StudentID).Return(exp, nil).Once()
		got, err := cq.GetStudent(ctx, nil, exp.StudentID)

		assert.NoError(t, err)
		assert.Equal(t, exp, got)
		m.AssertExpectations(t)

		hash, hErr := cache.HashString(prefStudent.String() + strconv.Itoa(int(exp.StudentID)))
		require.NoError(t, hErr)
		assert.Contains(t, c.Keys(), hash)
		assert.Equal(t, uint64(1), c.Len())
		assert.Equal(t, uint64(unsafe.Sizeof(exp))+uint64(len(exp.Data)), c.Size())
	})
}

func TestCachedQueriesGetTemplate(t *testing.T) {
	t.Run("if template cached, return it from cache and avoid db request", func(t *testing.T) {
		cq, _, m := prepCachedQueries(t)
		ctx := context.Background()
		exp := Template{
			TemplateID: 0,
			Content:    "",
		}
		m.EXPECT().CreateTemplate(ctx, nil, exp.Content).Return(exp, nil).Once()
		_, err := cq.CreateTemplate(ctx, nil, exp.Content)
		require.NoError(t, err)

		got, err := cq.GetTemplate(ctx, nil, exp.TemplateID)

		assert.NoError(t, err)
		assert.Equal(t, exp, got)
		m.AssertExpectations(t)
	})
	t.Run("if cache empty, make db request and cache result", func(t *testing.T) {
		cq, c, m := prepCachedQueries(t)
		ctx := context.Background()
		exp := Template{
			TemplateID: 0,
			Content:    "",
		}

		m.EXPECT().GetTemplate(ctx, nil, exp.TemplateID).Return(exp, nil).Once()
		got, err := cq.GetTemplate(ctx, nil, exp.TemplateID)

		assert.NoError(t, err)
		assert.Equal(t, exp, got)
		m.AssertExpectations(t)

		hash, hErr := cache.HashString(prefTmpl.String() + strconv.Itoa(int(exp.TemplateID)))
		require.NoError(t, hErr)
		assert.Contains(t, c.Keys(), hash)
		assert.Equal(t, uint64(1), c.Len())
		assert.Equal(t, uint64(unsafe.Sizeof(exp))+uint64(len(exp.Content)), c.Size())
	})
}

func TestCachedQueriesUpdateCertificate(t *testing.T) {
	t.Run("if certificate updated successfully it should be cached", func(t *testing.T) {
		cq, c, m := prepCachedQueries(t)
		ctx := context.Background()
		exp := Certificate{
			CertificateID: "00000000",
			TemplateID:    0,
			CourseID:      0,
			StudentID:     0,
			Timestamp: pgtype.Timestamptz{
				Time:             time.Now(),
				InfinityModifier: 0,
				Valid:            true,
			},
			Data: []byte{},
		}
		m.EXPECT().UpdateCertificate(ctx, nil, UpdateCertificateParams{}).Return(exp, nil).Once()

		got, err := cq.UpdateCertificate(ctx, nil, UpdateCertificateParams{})

		assert.NoError(t, err)
		assert.Equal(t, exp, got)
		m.AssertExpectations(t)

		hash, hErr := cache.HashString(prefCert.String() + exp.CertificateID)
		require.NoError(t, hErr)
		assert.Contains(t, c.Keys(), hash)
		assert.Equal(t, uint64(1), c.Len())
		assert.Equal(t, uint64(unsafe.Sizeof(exp))+uint64(len(exp.CertificateID)), c.Size())
	})
	t.Run("if certificate update failed cache should be intact", func(t *testing.T) {
		cq, c, m := prepCachedQueries(t)
		ctx := context.Background()
		cert := Certificate{
			CertificateID: "00000000",
			TemplateID:    0,
			CourseID:      0,
			StudentID:     0,
			Timestamp: pgtype.Timestamptz{
				Time:             time.Now(),
				InfinityModifier: 0,
				Valid:            true,
			},
			Data: []byte{},
		}
		params := UpdateCertificateParams{
			CertificateID: cert.CertificateID,
			Data:          []byte{},
		}
		m.EXPECT().CreateCertificate(ctx, nil, CreateCertificateParams{}).Return(cert, nil).Once()
		_, err := cq.CreateCertificate(ctx, nil, CreateCertificateParams{})
		require.NoError(t, err)
		require.Equal(t, uint64(1), c.Len())

		m.EXPECT().UpdateCertificate(ctx, nil, params).
			Return(Certificate{}, fmt.Errorf("failed")).Once()
		got, err := cq.UpdateCertificate(ctx, nil, params)

		assert.ErrorContains(t, err, "failed")
		assert.Empty(t, got)
		assert.Equal(t, uint64(1), c.Len())
		assert.Equal(t, uint64(unsafe.Sizeof(cert))+uint64(len(cert.CertificateID)), c.Size())
		m.AssertExpectations(t)
	})
}

func TestCachedQueriesUpdateCourse(t *testing.T) {
	t.Run("if course updated successfully it should be cached and linked certificates removed from cache", func(t *testing.T) {
		cq, c, m := prepCachedQueries(t)
		ctx := context.Background()
		exp := Course{
			CourseID: 0,
			Data:     []byte("{}"),
		}
		cert := Certificate{
			CertificateID: "00000000",
			TemplateID:    0,
			CourseID:      exp.CourseID,
			StudentID:     0,
			Timestamp: pgtype.Timestamptz{
				Time:             time.Now(),
				InfinityModifier: 0,
				Valid:            true,
			},
			Data: []byte{},
		}
		m.EXPECT().CreateCertificate(ctx, nil, CreateCertificateParams{}).Return(cert, nil).Once()
		_, err := cq.CreateCertificate(ctx, nil, CreateCertificateParams{})
		require.NoError(t, err)
		require.Equal(t, uint64(1), c.Len())

		m.EXPECT().UpdateCourse(ctx, nil, UpdateCourseParams{}).Return(exp, nil).Once()
		m.EXPECT().ListCertificatesByCourseLen(ctx, nil, exp.CourseID).Return(1, nil).Once()
		m.EXPECT().ListCertificatesByCourse(ctx, nil, ListCertificatesByCourseParams{
			CourseID: exp.CourseID,
			Limit:    1,
			Offset:   0,
		}).Return([]Certificate{cert}, nil).Once()
		got, err := cq.UpdateCourse(ctx, nil, UpdateCourseParams{})

		assert.NoError(t, err)
		assert.Equal(t, exp, got)
		m.AssertExpectations(t)

		hash, hErr := cache.HashString(prefCourse.String() + strconv.Itoa(int(exp.CourseID)))
		require.NoError(t, hErr)
		assert.Contains(t, c.Keys(), hash)
		assert.Equal(t, uint64(1), c.Len())
		assert.Equal(t, uint64(unsafe.Sizeof(exp))+uint64(len(exp.Data)), c.Size())
	})
	t.Run("if course update failed cache should be intact", func(t *testing.T) {
		cq, c, m := prepCachedQueries(t)
		ctx := context.Background()
		cert := Certificate{
			CertificateID: "00000000",
			TemplateID:    0,
			CourseID:      0,
			StudentID:     0,
			Timestamp: pgtype.Timestamptz{
				Time:             time.Now(),
				InfinityModifier: 0,
				Valid:            true,
			},
			Data: []byte{},
		}
		params := UpdateCourseParams{
			CourseID: cert.CourseID,
			Data:     []byte{},
		}
		m.EXPECT().CreateCertificate(ctx, nil, CreateCertificateParams{}).Return(cert, nil).Once()
		_, err := cq.CreateCertificate(ctx, nil, CreateCertificateParams{})
		require.NoError(t, err)
		require.Equal(t, uint64(1), c.Len())

		m.EXPECT().UpdateCourse(ctx, nil, params).Return(Course{}, fmt.Errorf("failed")).Once()

		got, err := cq.UpdateCourse(ctx, nil, params)

		assert.ErrorContains(t, err, "failed")
		assert.Empty(t, got)
		assert.Equal(t, uint64(1), c.Len())
		assert.Equal(t, uint64(unsafe.Sizeof(cert))+uint64(len(cert.CertificateID)), c.Size())
		m.AssertExpectations(t)
	})
}

func TestCachedQueriesUpdateStudent(t *testing.T) {
	t.Run("if student updated successfully it should be cached and linked certificates removed from cache", func(t *testing.T) {
		cq, c, m := prepCachedQueries(t)
		ctx := context.Background()
		exp := Student{
			StudentID: 0,
			Data:      []byte("{}"),
		}
		cert := Certificate{
			CertificateID: "00000000",
			TemplateID:    0,
			CourseID:      0,
			StudentID:     exp.StudentID,
			Timestamp: pgtype.Timestamptz{
				Time:             time.Now(),
				InfinityModifier: 0,
				Valid:            true,
			},
			Data: []byte{},
		}
		m.EXPECT().CreateCertificate(ctx, nil, CreateCertificateParams{}).Return(cert, nil).Once()
		_, err := cq.CreateCertificate(ctx, nil, CreateCertificateParams{})
		require.NoError(t, err)
		require.Equal(t, uint64(1), c.Len())

		m.EXPECT().UpdateStudent(ctx, nil, UpdateStudentParams{}).Return(exp, nil).Once()
		m.EXPECT().ListCertificatesByStudentLen(ctx, nil, exp.StudentID).Return(1, nil).Once()
		m.EXPECT().ListCertificatesByStudent(ctx, nil, ListCertificatesByStudentParams{
			StudentID: exp.StudentID,
			Limit:     1,
			Offset:    0,
		}).Return([]Certificate{cert}, nil).Once()
		got, err := cq.UpdateStudent(ctx, nil, UpdateStudentParams{})

		assert.NoError(t, err)
		assert.Equal(t, exp, got)
		m.AssertExpectations(t)

		hash, hErr := cache.HashString(prefStudent.String() + strconv.Itoa(int(exp.StudentID)))
		require.NoError(t, hErr)
		assert.Contains(t, c.Keys(), hash)
		assert.Equal(t, uint64(1), c.Len())
		assert.Equal(t, uint64(unsafe.Sizeof(exp))+uint64(len(exp.Data)), c.Size())
	})
	t.Run("if student update failed cache should be intact", func(t *testing.T) {
		cq, c, m := prepCachedQueries(t)
		ctx := context.Background()
		cert := Certificate{
			CertificateID: "00000000",
			TemplateID:    0,
			CourseID:      0,
			StudentID:     0,
			Timestamp: pgtype.Timestamptz{
				Time:             time.Now(),
				InfinityModifier: 0,
				Valid:            true,
			},
			Data: []byte{},
		}
		params := UpdateStudentParams{
			StudentID: cert.StudentID,
			Data:      []byte{},
		}
		m.EXPECT().CreateCertificate(ctx, nil, CreateCertificateParams{}).Return(cert, nil).Once()
		_, err := cq.CreateCertificate(ctx, nil, CreateCertificateParams{})
		require.NoError(t, err)
		require.Equal(t, uint64(1), c.Len())

		m.EXPECT().UpdateStudent(ctx, nil, params).Return(Student{}, fmt.Errorf("failed")).Once()

		got, err := cq.UpdateStudent(ctx, nil, params)

		assert.ErrorContains(t, err, "failed")
		assert.Empty(t, got)
		assert.Equal(t, uint64(1), c.Len())
		assert.Equal(t, uint64(unsafe.Sizeof(cert))+uint64(len(cert.CertificateID)), c.Size())
		m.AssertExpectations(t)
	})
}

func TestCachedQueriesUpdateTemplate(t *testing.T) {
	t.Run("if template updated successfully it should be cached and linked certificates removed from cache", func(t *testing.T) {
		cq, c, m := prepCachedQueries(t)
		ctx := context.Background()
		exp := Template{
			TemplateID: 0,
			Content:    "",
		}
		cert := Certificate{
			CertificateID: "00000000",
			TemplateID:    exp.TemplateID,
			CourseID:      0,
			StudentID:     0,
			Timestamp: pgtype.Timestamptz{
				Time:             time.Now(),
				InfinityModifier: 0,
				Valid:            true,
			},
			Data: []byte{},
		}
		m.EXPECT().CreateCertificate(ctx, nil, CreateCertificateParams{}).Return(cert, nil).Once()
		_, err := cq.CreateCertificate(ctx, nil, CreateCertificateParams{})
		require.NoError(t, err)
		require.Equal(t, uint64(1), c.Len())

		m.EXPECT().UpdateTemplate(ctx, nil, UpdateTemplateParams{}).Return(exp, nil).Once()
		m.EXPECT().ListCertificatesByTemplateLen(ctx, nil, exp.TemplateID).Return(1, nil).Once()
		m.EXPECT().ListCertificatesByTemplate(ctx, nil, ListCertificatesByTemplateParams{
			TemplateID: exp.TemplateID,
			Limit:      1,
			Offset:     0,
		}).Return([]Certificate{cert}, nil).Once()
		got, err := cq.UpdateTemplate(ctx, nil, UpdateTemplateParams{})

		assert.NoError(t, err)
		assert.Equal(t, exp, got)
		m.AssertExpectations(t)

		hash, hErr := cache.HashString(prefTmpl.String() + strconv.Itoa(int(exp.TemplateID)))
		require.NoError(t, hErr)
		assert.Contains(t, c.Keys(), hash)
		assert.Equal(t, uint64(1), c.Len())
		assert.Equal(t, uint64(unsafe.Sizeof(exp))+uint64(len(exp.Content)), c.Size())
	})
	t.Run("if student update failed cache should be intact", func(t *testing.T) {
		cq, c, m := prepCachedQueries(t)
		ctx := context.Background()
		cert := Certificate{
			CertificateID: "00000000",
			TemplateID:    0,
			CourseID:      0,
			StudentID:     0,
			Timestamp: pgtype.Timestamptz{
				Time:             time.Now(),
				InfinityModifier: 0,
				Valid:            true,
			},
			Data: []byte{},
		}
		params := UpdateTemplateParams{
			TemplateID: cert.TemplateID,
			Content:    "",
		}
		m.EXPECT().CreateCertificate(ctx, nil, CreateCertificateParams{}).Return(cert, nil).Once()
		_, err := cq.CreateCertificate(ctx, nil, CreateCertificateParams{})
		require.NoError(t, err)
		require.Equal(t, uint64(1), c.Len())

		m.EXPECT().UpdateTemplate(ctx, nil, params).Return(Template{}, fmt.Errorf("failed")).Once()

		got, err := cq.UpdateTemplate(ctx, nil, params)

		assert.ErrorContains(t, err, "failed")
		assert.Empty(t, got)
		assert.Equal(t, uint64(1), c.Len())
		assert.Equal(t, uint64(unsafe.Sizeof(cert))+uint64(len(cert.CertificateID)), c.Size())
		m.AssertExpectations(t)
	})
}
