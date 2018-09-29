package oaichecker

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-openapi/analysis"
)

type Analyzer struct {
	analyzer *analysis.Spec
}

func NewAnalyzer(specs *Specs) *Analyzer {
	if specs == nil {
		return &Analyzer{}
	}

	return &Analyzer{
		analyzer: analysis.New(specs.document.OrigSpec()),
	}
}

func (t *Analyzer) Analyze(req *http.Request) error {
	if t.analyzer == nil {
		return errors.New("no specs defined")
	}

	if req == nil {
		return errors.New("no request defined")
	}

	operation, ok := t.analyzer.OperationFor(req.Method, req.URL.Path)
	if !ok {
		return errors.New("operation not defined inside the specs")
	}

	fmt.Printf("operation: %#v\n", operation)

	return nil
}
