package oaichecker

import "net/http"

type Transport struct {
	Transport http.RoundTripper
}

func NewTransport() *Transport {
	return &Transport{
		Transport: http.DefaultTransport,
	}
}

func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	res, err := t.Transport.RoundTrip(req)

	return res, err
}
