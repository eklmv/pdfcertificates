package render

import (
	"io"
)

type Renderer interface {
	getStage() stage
	setStage(s stage)
	nextStage() stage
	isValidStage() bool
	Render(in io.Reader, out io.Writer, data *Data) error
}

type stage int

const (
	prep = iota
	html
	pdf
)

func (s stage) String() string {
	return []string{"prep", "html", "pdf"}[s]
}
