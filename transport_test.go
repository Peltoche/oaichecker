package oaichecker

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockTransport struct {
	mock.Mock
}

func (t *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	args := t.Called(req)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*http.Response), args.Error(1)
}

func resBody(t *testing.T, res *http.Response) string {
	body, err := ioutil.ReadAll(res.Body)
	require.NoError(t, err)

	return string(body)
}

func newServer(handlerFunc func(http.ResponseWriter, *http.Request)) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(handlerFunc))
}

func Test_Transport_implements_RoundTripper(t *testing.T) {
	assert.Implements(t, (*http.RoundTripper)(nil), &Transport{})
}

func Test_Transport_with_a_valid_request(t *testing.T) {
	ts := newServer(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("some-response"))
		require.NoError(t, err)
	})
	defer ts.Close()

	specs, err := NewSpecsFromFile("./dataset/petstore_minimal.json")
	require.NoError(t, err)

	client := http.Client{
		Transport: NewTransport(specs),
	}

	res, err := client.Get(ts.URL + "/pets")

	assert.NoError(t, err)
	assert.Equal(t, "some-response", resBody(t, res))
}

func Test_Transport_with_a_transport_error(t *testing.T) {
	specs, err := NewSpecsFromFile("./dataset/petstore_minimal.json")
	require.NoError(t, err)

	mockInnerTransport := new(mockTransport)
	mockInnerTransport.On("RoundTrip", mock.Anything).Return(nil, errors.New("some-error")).Once()

	checkerTransport := NewTransport(specs)
	checkerTransport.Transport = mockInnerTransport

	client := http.Client{
		Transport: checkerTransport,
	}

	res, err := client.Get("http://foobar/pets")

	assert.Nil(t, res)
	assert.EqualError(t, err, "Get http://foobar/pets: some-error")

	mockInnerTransport.AssertExpectations(t)
}

func Test_Transport_with_an_analyzer_error(t *testing.T) {
	ts := newServer(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("some-response"))
		require.NoError(t, err)
	})
	defer ts.Close()

	specs, err := NewSpecsFromFile("./dataset/petstore_minimal.json")
	require.NoError(t, err)

	client := http.Client{
		Transport: NewTransport(specs),
	}

	res, err := client.Get(ts.URL + "/invalid-path")

	assert.Nil(t, res)
	assert.EqualError(t, err, fmt.Sprintf("Get %s/invalid-path: operation not defined inside the specs", ts.URL))
}

func Test_Transport_with_a_body(t *testing.T) {
	ts := newServer(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("some-response"))
		require.NoError(t, err)
	})
	defer ts.Close()

	specs, err := NewSpecsFromFile("./dataset/petstore.json")
	require.NoError(t, err)

	client := http.Client{
		Transport: NewTransport(specs),
	}

	res, err := client.Post(ts.URL+"/pet", "application/json", strings.NewReader(`{
		"name": "foobar",
		"photoUrls": ["some-url"]
	}`))

	assert.NoError(t, err)
	assert.Equal(t, "some-response", resBody(t, res))
}
