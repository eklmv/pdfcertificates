package render

import (
	"bytes"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
)

const route = "/forms/chromium/convert/html"

type GotenbergRender struct {
	s   stage
	url string
}

func NewGotenbergRender(url string) *GotenbergRender {
	return &GotenbergRender{url: url}
}

func (g *GotenbergRender) getStage() stage {
	return g.s
}

func (g *GotenbergRender) setStage(s stage) {
	g.s = s
}

func (g *GotenbergRender) nextStage() stage {
	return pdf
}

func (g *GotenbergRender) isValidStage() bool {
	return g.s == html
}

func (g *GotenbergRender) Render(in io.Reader, out io.Writer, data *Data) error {
	body := new(bytes.Buffer)
	wr := multipart.NewWriter(body)
	ff, err := wr.CreateFormFile("files", "index.html")
	if err != nil {
		slog.Error("failed to create form file", slog.Any("error", err))
		return err
	}
	b, err := io.ReadAll(in)
	if err != nil {
		slog.Error("failed to read rendered html", slog.Any("in", in), slog.Any("error", err))
		return err
	}
	_, err = ff.Write(b)
	if err != nil {
		slog.Error("failed to write rendered html file to form", slog.Any("error", err))
		return err
	}
	err = wr.WriteField("preferCssPageSize", "true")
	if err != nil {
		slog.Error("failed to write form field", slog.Any("error", err))
		return err
	}
	err = wr.Close()
	if err != nil {
		slog.Error("failed to close multipart message", slog.Any("error", err))
		return err
	}
	resp, err := http.Post(g.url+route, wr.FormDataContentType(), body)
	if err != nil {
		slog.Error("failed to perform POST request to gotenberg service", slog.Any("error", err))
		return err
	}
	pdf, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("failed to read responce body", slog.Any("error", err))
		return err
	}
	if resp.StatusCode != http.StatusOK {
		slog.Error("failed to generate pdf", slog.Any("error", string(pdf)))
	}
	_, err = out.Write(pdf)
	if err != nil {
		slog.Error("failed to write to out", slog.Any("out", out), slog.Any("error", err))
	}
	return nil
}
