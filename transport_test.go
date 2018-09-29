package oaichecker

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func resBody(t *testing.T, res *http.Response) string {
	body, err := ioutil.ReadAll(res.Body)
	require.NoError(t, err)

	return string(body)
}

func newServer(handlerFunc func(http.ResponseWriter, *http.Request)) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(handlerFunc))
}

func Test_Transport_implements_RoundTripper(t *testing.T) {
	assert.Implements(t, (*http.RoundTripper)(nil), NewTransport())
}

func Test_Transport_emit_the_request(t *testing.T) {
	ts := newServer(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("some-response"))
	})
	defer ts.Close()

	client := http.Client{
		Transport: NewTransport(),
	}

	res, err := client.Get(ts.URL)

	assert.NoError(t, err)
	assert.Equal(t, "some-response", resBody(t, res))
}
