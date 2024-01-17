package render

import (
	"encoding/json"
	"log/slog"

	"github.com/eklmv/pdfcertificates/internal/db"
)

type Data struct {
	CertificateID string
	Link          string
	Certificate   map[string]any
	Course        map[string]any
	Student       map[string]any
}

func ExtractData(host string, cert db.Certificate, course db.Course, student db.Student) (data Data, err error) {
	data.CertificateID = cert.CertificateID
	data.Link = host + data.CertificateID
	err = json.Unmarshal(cert.Data, &data.Certificate)
	if err != nil {
		slog.Error("failed to unmarshal certificate's data", slog.Any("data", cert.Data), slog.Any("error", err))
		return
	}
	err = json.Unmarshal(course.Data, &data.Course)
	if err != nil {
		slog.Error("failed to unmarshal course's data", slog.Any("data", course.Data), slog.Any("error", err))
		return
	}
	err = json.Unmarshal(student.Data, &data.Student)
	if err != nil {
		slog.Error("failed to unmarshal student's data", slog.Any("data", student.Data), slog.Any("error", err))
		return
	}
	return
}
