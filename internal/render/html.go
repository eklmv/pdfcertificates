package render

import (
	"html/template"
	"io"
	"log/slog"
)

type HTMLRender struct {
	s stage
}

func (h *HTMLRender) getStage() stage {
	return h.s
}

func (h *HTMLRender) setStage(s stage) {
	h.s = s
}

func (h *HTMLRender) nextStage() stage {
	return html
}

func (h *HTMLRender) isValidStage() bool {
	return h.s == prep
}

func (*HTMLRender) Render(in io.Reader, out io.Writer, data *Data) error {
	source, err := io.ReadAll(in)
	if err != nil {
		slog.Error("failed to read html template", slog.Any("in", in), slog.Any("error", err))
		return err
	}

	tmpl, err := template.New("HTML").Parse(string(source))
	if err != nil {
		slog.Error("failed to parse html template", slog.Any("template", tmpl), slog.Any("error", err))
		return err
	}

	err = tmpl.Execute(out, data)
	if err != nil {
		slog.Error("failed to execute html template",
			slog.Any("out", out), slog.Any("data", data), slog.Any("error", err))
		return err
	}

	return nil
}
