package render

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTMLRenderImplementsInterface(t *testing.T) {
	assert.Implements(t, (*Renderer)(nil), new(HTMLRender))
}

func TestHTMLRender(t *testing.T) {
	data := Data{
		CertificateID: "00000000",
		Link:          "localhost",
		Certificate:   map[string]any{"issued": "19 Dec 1999"},
		Course:        map[string]any{"title": "Test Course"},
		Student:       map[string]any{"name": "Test Student"},
	}
	template := `
	<p>{{.CertificateID}}</p>
	<p>{{.Link}}</p>
	<p>{{.Certificate.issued}}</p>
	<p>{{.Course.title}}</p>
	<p>{{.Student.name}}</p>
	`
	exp := `
	<p>` + data.CertificateID + `</p>
	<p>` + data.Link + `</p>
	<p>` + data.Certificate["issued"].(string) + `</p>
	<p>` + data.Course["title"].(string) + `</p>
	<p>` + data.Student["name"].(string) + `</p>
	`

	got := strings.Builder{}
	err := new(HTMLRender).Render(strings.NewReader(template), &got, &data)

	require.NoError(t, err)
	assert.Equal(t, exp, got.String())
}
