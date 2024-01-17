package db

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/eklmv/pdfcertificates/internal/cache"
)

type CachedQueries struct {
	Querier
	c cache.Cache[uint32, cachedResponse]
}

type cachedResponse struct {
	value any
	size  uint64
}

func (r cachedResponse) Size() uint64 {
	return r.size
}

type prefix int

const (
	CERTIFICATE prefix = iota
	COURSE
	STUDENT
	TEMPLATE
)

func (p prefix) String() string {
	return []string{"certificate_", "course_", "student_", "template_"}[p]
}

func NewCachedQueries(cache cache.Cache[uint32, cachedResponse], querier Querier) *CachedQueries {
	return &CachedQueries{
		Querier: querier,
		c:       cache,
	}
}

func (cq *CachedQueries) addToCache(p prefix, str string, value any) {
	hash, err := cache.HashString(p.String() + str)
	if err != nil {
		slog.Error("skip caching", slog.String("prefix", p.String()),
			slog.String("str", str), slog.Any("error", err))
		return
	}
	cq.c.Add(hash, cachedResponse{
		value: value,
		size:  cache.SizeOf(value),
	})
}

func (cq *CachedQueries) invalidateCache(p prefix, str string) {
	hash, err := cache.HashString(p.String() + str)
	if err == nil && cq.c.Contains(hash) {
		cq.c.Remove(hash)
	}
	if err != nil {
		slog.Error("cache can be invalid, force purge", slog.Any("error", err))
		cq.c.Purge()
	}
}

// TODO: rewrite using some sort of thread safe lookup index
func (cq *CachedQueries) invalidateCertificates(ctx context.Context, db DBTX, p prefix, id int32) {
	var err error
	var l int64
	var certs []Certificate
	switch p {
	case TEMPLATE:
		l, err = cq.ListCertificatesByTemplateLen(ctx, db, id)
		if err != nil {
			break
		}
		certs, err = cq.ListCertificatesByTemplate(ctx, db, ListCertificatesByTemplateParams{
			TemplateID: id,
			Limit:      l,
			Offset:     0,
		})
	case COURSE:
		l, err = cq.ListCertificatesByCourseLen(ctx, db, id)
		if err != nil {
			break
		}
		certs, err = cq.ListCertificatesByCourse(ctx, db, ListCertificatesByCourseParams{
			CourseID: id,
			Limit:    l,
			Offset:   0,
		})
	case STUDENT:
		l, err = cq.ListCertificatesByStudentLen(ctx, db, id)
		if err != nil {
			break
		}
		certs, err = cq.ListCertificatesByStudent(ctx, db, ListCertificatesByStudentParams{
			StudentID: id,
			Limit:     l,
			Offset:    0,
		})
	default:
		err = fmt.Errorf("invalid prefix")
	}
	if err != nil {
		slog.Error("failed to invalidate cached certificates, force purge",
			slog.String("prefix", p.String()), slog.Any("error", err))
		cq.c.Purge()
		return
	}
	for _, c := range certs {
		cq.invalidateCache(CERTIFICATE, c.CertificateID)
	}
}

func (cq *CachedQueries) hitCache(p prefix, str string) (v any, ok bool) {
	hash, err := cache.HashString(p.String() + str)
	if err != nil {
		slog.Error("skip cache hit", slog.Any("error", err))
		return
	}
	if cq.c.Contains(hash) {
		r, ok := cq.c.Get(hash)
		slog.Info("successful cache hit", slog.String("hashed string", p.String()+str),
			slog.Any("value", r.value), slog.Bool("ok", ok))
		return r.value, ok
	}
	return
}

func (cq *CachedQueries) CreateCertificate(ctx context.Context, db DBTX, arg CreateCertificateParams) (Certificate, error) {
	cert, err := cq.Querier.CreateCertificate(ctx, db, arg)
	if err == nil {
		cq.addToCache(CERTIFICATE, cert.CertificateID, cert)
	}
	return cert, err
}

func (cq *CachedQueries) CreateCourse(ctx context.Context, db DBTX, data []byte) (Course, error) {
	course, err := cq.Querier.CreateCourse(ctx, db, data)
	if err == nil {
		cq.addToCache(COURSE, strconv.Itoa(int(course.CourseID)), course)
	}
	return course, err
}

func (cq *CachedQueries) CreateStudent(ctx context.Context, db DBTX, data []byte) (Student, error) {
	student, err := cq.Querier.CreateStudent(ctx, db, data)
	if err == nil {
		cq.addToCache(STUDENT, strconv.Itoa(int(student.StudentID)), student)
	}
	return student, err
}

func (cq *CachedQueries) CreateTemplate(ctx context.Context, db DBTX, content string) (Template, error) {
	tmpl, err := cq.Querier.CreateTemplate(ctx, db, content)
	if err == nil {
		cq.addToCache(TEMPLATE, strconv.Itoa(int(tmpl.TemplateID)), tmpl)
	}
	return tmpl, err
}

func (cq *CachedQueries) DeleteCertificate(ctx context.Context, db DBTX, certificateID string) (Certificate, error) {
	cert, err := cq.Querier.DeleteCertificate(ctx, db, certificateID)
	if err == nil {
		cq.invalidateCache(CERTIFICATE, certificateID)
	}
	return cert, err
}

func (cq *CachedQueries) DeleteCourse(ctx context.Context, db DBTX, courseID int32) (Course, error) {
	course, err := cq.Querier.DeleteCourse(ctx, db, courseID)
	if err == nil {
		cq.invalidateCache(COURSE, strconv.Itoa(int(courseID)))
	}
	return course, err
}

func (cq *CachedQueries) DeleteStudent(ctx context.Context, db DBTX, studentID int32) (Student, error) {
	student, err := cq.Querier.DeleteStudent(ctx, db, studentID)
	if err == nil {
		cq.invalidateCache(STUDENT, strconv.Itoa(int(studentID)))
		cq.invalidateCertificates(ctx, db, STUDENT, studentID)
	}
	return student, err
}

func (cq *CachedQueries) DeleteTemplate(ctx context.Context, db DBTX, templateID int32) (Template, error) {
	tmpl, err := cq.Querier.DeleteTemplate(ctx, db, templateID)
	if err == nil {
		cq.invalidateCache(TEMPLATE, strconv.Itoa(int(templateID)))
	}
	return tmpl, err
}

func (cq *CachedQueries) GetCertificate(ctx context.Context, db DBTX, certificateID string) (Certificate, error) {
	if v, ok := cq.hitCache(CERTIFICATE, certificateID); ok {
		if cert, ok := v.(Certificate); ok {
			return cert, nil
		} else {
			slog.Error("failed type conversion of cached value", slog.String("scope", "GetCertificate"),
				slog.String("type", "Certificate"), slog.Any("value", v))
			cq.invalidateCache(CERTIFICATE, certificateID)
		}
	}

	cert, err := cq.Querier.GetCertificate(ctx, db, certificateID)
	if err == nil {
		cq.addToCache(CERTIFICATE, certificateID, cert)
	}

	return cert, err
}

func (cq *CachedQueries) GetCourse(ctx context.Context, db DBTX, courseID int32) (Course, error) {
	if v, ok := cq.hitCache(COURSE, strconv.Itoa(int(courseID))); ok {
		if course, ok := v.(Course); ok {
			return course, nil
		} else {
			slog.Error("failed type conversion of cached value", slog.String("scope", "GetCourse"),
				slog.String("type", "Course"), slog.Any("value", v))
			cq.invalidateCache(COURSE, strconv.Itoa(int(courseID)))
		}
	}
	course, err := cq.Querier.GetCourse(ctx, db, courseID)
	if err == nil {
		cq.addToCache(COURSE, strconv.Itoa(int(courseID)), course)
	}

	return course, err
}

func (cq *CachedQueries) GetStudent(ctx context.Context, db DBTX, studentID int32) (Student, error) {
	if v, ok := cq.hitCache(STUDENT, strconv.Itoa(int(studentID))); ok {
		if student, ok := v.(Student); ok {
			return student, nil
		} else {
			slog.Error("failed type conversion of cached value", slog.String("scope", "GetStudent"),
				slog.String("type", "Student"), slog.Any("value", v))
			cq.invalidateCache(STUDENT, strconv.Itoa(int(studentID)))
		}
	}

	student, err := cq.Querier.GetStudent(ctx, db, studentID)
	if err == nil {
		cq.addToCache(STUDENT, strconv.Itoa(int(studentID)), student)
	}

	return student, err
}

func (cq *CachedQueries) GetTemplate(ctx context.Context, db DBTX, templateID int32) (Template, error) {
	if v, ok := cq.hitCache(TEMPLATE, strconv.Itoa(int(templateID))); ok {
		if tmpl, ok := v.(Template); ok {
			return tmpl, nil
		} else {
			slog.Error("failed type conversion of cached value", slog.String("scope", "GetTemplate"),
				slog.String("type", "Template"), slog.Any("value", v))
			cq.invalidateCache(TEMPLATE, strconv.Itoa(int(templateID)))
		}
	}

	tmpl, err := cq.Querier.GetTemplate(ctx, db, templateID)
	if err == nil {
		cq.addToCache(TEMPLATE, strconv.Itoa(int(templateID)), tmpl)
	}

	return tmpl, err
}

func (cq *CachedQueries) UpdateCertificate(ctx context.Context, db DBTX, arg UpdateCertificateParams) (Certificate, error) {
	cert, err := cq.Querier.UpdateCertificate(ctx, db, arg)
	if err == nil {
		cq.addToCache(CERTIFICATE, cert.CertificateID, cert)
	}
	return cert, err
}

func (cq *CachedQueries) UpdateCourse(ctx context.Context, db DBTX, arg UpdateCourseParams) (Course, error) {
	course, err := cq.Querier.UpdateCourse(ctx, db, arg)
	if err == nil {
		cq.addToCache(COURSE, strconv.Itoa(int(arg.CourseID)), course)
		cq.invalidateCertificates(ctx, db, COURSE, arg.CourseID)
	}
	return course, err
}
func (cq *CachedQueries) UpdateStudent(ctx context.Context, db DBTX, arg UpdateStudentParams) (Student, error) {
	student, err := cq.Querier.UpdateStudent(ctx, db, arg)
	if err == nil {
		cq.addToCache(STUDENT, strconv.Itoa(int(arg.StudentID)), student)
		cq.invalidateCertificates(ctx, db, STUDENT, arg.StudentID)
	}
	return student, err
}
func (cq *CachedQueries) UpdateTemplate(ctx context.Context, db DBTX, arg UpdateTemplateParams) (Template, error) {
	tmpl, err := cq.Querier.UpdateTemplate(ctx, db, arg)
	if err == nil {
		cq.addToCache(TEMPLATE, strconv.Itoa(int(arg.TemplateID)), tmpl)
		cq.invalidateCertificates(ctx, db, TEMPLATE, arg.TemplateID)
	}
	return tmpl, err
}
