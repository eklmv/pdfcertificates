//go:build integration

package render

import (
	"bytes"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGotenbergRenderImplementsInterface(t *testing.T) {
	assert.Implements(t, (*Renderer)(nil), new(GotenbergRender))
}

func TestGotenbergRender(t *testing.T) {
	host := os.Getenv("GOTENBERG_TEST_IP")
	port := os.Getenv("GOTENBERG_TEST_PORT")
	require.NotEmpty(t, host)
	require.NotEmpty(t, port)
	testData := "../../test/testdata/integration/gotenberg/"
	html, err := os.ReadFile(testData + "in.html")
	require.NoError(t, err)
	golden, err := os.ReadFile(testData + "out.pdf")
	require.NoError(t, err)
	dates := regexp.MustCompile("/CreationDate.*\n/ModDate.*\n")
	exp := dates.ReplaceAllString(string(golden), "")

	g := NewGotenbergRender("http://" + host + ":" + port)
	in := bytes.NewReader(html)
	out := new(strings.Builder)

	err = g.Render(in, out, nil)

	require.NoError(t, err)
	got := dates.ReplaceAllString(out.String(), "")
	assert.NotEmpty(t, got)
	assert.Equal(t, exp, got)
}
