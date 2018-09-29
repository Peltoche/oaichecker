package oaichecker

import "net/http"

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
