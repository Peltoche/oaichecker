package oaichecker

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
)

// Transport is a http.RoundTripper implementation destined to be injected
// inside an htttp.Client.Transport.
//
// It allows to make HTTP calls as usual but it will intercept the
// http.Request and http.Response and will validate them against the OpenAPI
// specs given during instantiation.
type Transport struct {
	Transport http.RoundTripper
	analyzer  *Analyzer
}

// NewTransport instantiate a new Transport with the given Specs.
func NewTransport(specs *Specs) *Transport {
	return &Transport{
		Transport: http.DefaultTransport,
		analyzer:  NewAnalyzer(specs),
	}
}

// RoundTrip implement http.RoundTripper.
//
// If a validation error occures an error will returned with a new Response.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	var (
		err  error
		body []byte
	)

	// GetBody is an optional func to return a new copy of Body
	switch req.Body.(type) {
	case nil:
		req.GetBody = func() (io.ReadCloser, error) {
			return http.NoBody, nil
		}
	default:
		body, err = ioutil.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}

		req.GetBody = func() (io.ReadCloser, error) {
			return ioutil.NopCloser(bytes.NewReader(body)), nil
		}
	}

	req.Body, err = req.GetBody()
	if err != nil {
		return nil, err
	}

	res, err := t.Transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	err = t.analyzer.Analyze(req, res)
	if err != nil {
		return nil, err
	}

	return res, err
}
