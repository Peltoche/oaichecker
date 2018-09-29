package oaichecker

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
)

type Transport struct {
	Transport http.RoundTripper
	analyzer  *Analyzer
}

func NewTransport(specs *Specs) *Transport {
	return &Transport{
		Transport: http.DefaultTransport,
		analyzer:  NewAnalyzer(specs),
	}
}

func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	var err error

	// GetBody is an optional func to return a new copy of Body
	switch req.Body.(type) {
	case nil:
		req.GetBody = func() (io.ReadCloser, error) {
			return http.NoBody, nil
		}
	default:
		body, err := ioutil.ReadAll(req.Body)
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

	err = t.analyzer.Analyze(req)
	if err != nil {
		return nil, err
	}

	return res, err
}
