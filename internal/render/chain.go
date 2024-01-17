package render

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
)

type ChainRender struct {
	s     stage
	chain []Renderer
}

func (c *ChainRender) getStage() stage {
	return c.s
}

func (c *ChainRender) setStage(s stage) {
	c.s = s
}

func (c *ChainRender) nextStage() stage {
	return c.s
}

func (c *ChainRender) isValidStage() bool {
	return true
}

func (c *ChainRender) Render(in io.Reader, out io.Writer, data *Data) error {
	buf := new(bytes.Buffer)
	tmpIn := in
	tmpOut := buf
	for _, r := range c.chain {
		r.setStage(c.getStage())
		if !r.isValidStage() {
			err := fmt.Errorf("invalid stage: %s", c.getStage())
			slog.Error("failed to execute step in render chain", slog.Any("error", err))
			return err
		}
		err := r.Render(tmpIn, tmpOut, data)
		if err != nil {
			slog.Error("failed to execute step in render chain", slog.Any("error", err))
			return err
		}
		c.setStage(r.nextStage())
		buf = tmpOut
		tmpIn = tmpOut
		tmpOut = new(bytes.Buffer)
	}
	_, err := buf.WriteTo(out)
	if err != nil {
		slog.Error("failed to execute render chain", slog.Any("error", err))
		return err
	}
	return nil
}

func (c *ChainRender) Append(r Renderer) *ChainRender {
	c.chain = append(c.chain, r)
	return c
}
