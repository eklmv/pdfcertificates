// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.24.0

package db

import (
	"context"
)

type Querier interface {
	CreateCertificate(ctx context.Context, db DBTX, arg CreateCertificateParams) (Certificate, error)
	CreateCourse(ctx context.Context, db DBTX, data []byte) (Course, error)
	CreateStudent(ctx context.Context, db DBTX, data []byte) (Student, error)
	CreateTemplate(ctx context.Context, db DBTX, content string) (Template, error)
	DeleteCertificate(ctx context.Context, db DBTX, certificateID string) (Certificate, error)
	DeleteCourse(ctx context.Context, db DBTX, courseID int32) (Course, error)
	DeleteStudent(ctx context.Context, db DBTX, studentID int32) (Student, error)
	DeleteTemplate(ctx context.Context, db DBTX, templateID int32) (Template, error)
	GetCertificate(ctx context.Context, db DBTX, certificateID string) (Certificate, error)
	GetCourse(ctx context.Context, db DBTX, courseID int32) (Course, error)
	GetStudent(ctx context.Context, db DBTX, studentID int32) (Student, error)
	GetTemplate(ctx context.Context, db DBTX, templateID int32) (Template, error)
	ListCertificates(ctx context.Context, db DBTX, arg ListCertificatesParams) ([]Certificate, error)
	ListCertificatesByCourse(ctx context.Context, db DBTX, arg ListCertificatesByCourseParams) ([]Certificate, error)
	ListCertificatesByCourseLen(ctx context.Context, db DBTX, courseID int32) (int64, error)
	ListCertificatesByStudent(ctx context.Context, db DBTX, arg ListCertificatesByStudentParams) ([]Certificate, error)
	ListCertificatesByStudentLen(ctx context.Context, db DBTX, studentID int32) (int64, error)
	ListCertificatesByTemplate(ctx context.Context, db DBTX, arg ListCertificatesByTemplateParams) ([]Certificate, error)
	ListCertificatesByTemplateLen(ctx context.Context, db DBTX, templateID int32) (int64, error)
	ListCertificatesLen(ctx context.Context, db DBTX) (int64, error)
	ListCourses(ctx context.Context, db DBTX, arg ListCoursesParams) ([]Course, error)
	ListCoursesLen(ctx context.Context, db DBTX) (int64, error)
	ListStudents(ctx context.Context, db DBTX, arg ListStudentsParams) ([]Student, error)
	ListStudentsLen(ctx context.Context, db DBTX) (int64, error)
	ListTemplates(ctx context.Context, db DBTX, arg ListTemplatesParams) ([]Template, error)
	ListTemplatesLen(ctx context.Context, db DBTX) (int64, error)
	UpdateCertificate(ctx context.Context, db DBTX, arg UpdateCertificateParams) (Certificate, error)
	UpdateCourse(ctx context.Context, db DBTX, arg UpdateCourseParams) (Course, error)
	UpdateStudent(ctx context.Context, db DBTX, arg UpdateStudentParams) (Student, error)
	UpdateTemplate(ctx context.Context, db DBTX, arg UpdateTemplateParams) (Template, error)
}

var _ Querier = (*Queries)(nil)
