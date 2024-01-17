package render

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChainRenderImplementsInterface(t *testing.T) {
	assert.Implements(t, (*Renderer)(nil), new(ChainRender))
}

type mockRender struct {
	t          *testing.T
	s          stage
	nextS      stage
	validFunc  func(s stage) bool
	mockRender func(t *testing.T, in io.Reader, out io.Writer, data *Data)
	err        error
}

func (m *mockRender) getStage() stage {
	return m.s
}

func (m *mockRender) setStage(s stage) {
	m.s = s
}

func (m *mockRender) nextStage() stage {
	return m.nextS
}

func (m *mockRender) isValidStage() bool {
	return m.validFunc(m.s)
}

func (m *mockRender) Render(in io.Reader, out io.Writer, data *Data) error {
	if m.mockRender != nil {
		m.mockRender(m.t, in, out, data)
	}
	return m.err
}

func TestChainRender(t *testing.T) {
	t.Run("properly propagate stage and in, out, data arguments", func(t *testing.T) {
		expIn := "in"
		callIndex := 0
		expOut1 := "out1"
		expOut2 := "out2"
		expOut3 := "out3"
		expData := Data{
			Link: "test",
		}
		exp := expIn + expOut1 + expOut2 + expOut3
		m1 := &mockRender{
			t:     t,
			nextS: html,
			validFunc: func(s stage) bool {
				return s <= prep
			},
			mockRender: func(t *testing.T, in io.Reader, out io.Writer, data *Data) {
				b, err := io.ReadAll(in)
				require.NoError(t, err)
				assert.Equal(t, expIn, string(b))

				_, err = out.Write(append(b, []byte(expOut1)...))
				require.NoError(t, err)

				data.Link = expData.Link
				assert.Equal(t, expData, *data)

				assert.Equal(t, 0, callIndex)
				callIndex++
			},
			err: nil,
		}
		m2 := &mockRender{
			t:     t,
			nextS: pdf,
			validFunc: func(s stage) bool {
				return s == html
			},
			mockRender: func(t *testing.T, in io.Reader, out io.Writer, data *Data) {
				b, err := io.ReadAll(in)
				require.NoError(t, err)
				assert.Equal(t, expIn+expOut1, string(b))

				_, err = out.Write(append(b, []byte(expOut2)...))
				require.NoError(t, err)

				assert.Equal(t, expData, *data)

				assert.Equal(t, 1, callIndex)
				callIndex++
			},
			err: nil,
		}
		m3 := &mockRender{
			t:     t,
			nextS: pdf,
			validFunc: func(s stage) bool {
				return s == pdf
			},
			mockRender: func(t *testing.T, in io.Reader, out io.Writer, data *Data) {
				b, err := io.ReadAll(in)
				require.NoError(t, err)
				assert.Equal(t, expIn+expOut1+expOut2, string(b))

				_, err = out.Write(append(b, []byte(expOut3)...))
				require.NoError(t, err)

				assert.Equal(t, expData, *data)

				assert.Equal(t, 2, callIndex)
				callIndex++
			},
			err: nil,
		}

		in := strings.Clone(expIn)
		out := new(strings.Builder)
		data := new(Data)
		innerChain := new(ChainRender).Append(m2).Append(m3)
		err := new(ChainRender).Append(m1).Append(innerChain).Render(strings.NewReader(in), out, data)

		require.NoError(t, err)
		assert.Equal(t, expIn, in)
		assert.Equal(t, expData, *data)
		assert.Equal(t, exp, out.String())
	})
}
