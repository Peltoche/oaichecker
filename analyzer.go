package oaichecker

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-openapi/analysis"
	"github.com/go-openapi/spec"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"
)

type Analyzer struct {
	analyzer *analysis.Spec
	schema   *spec.Schema
}

func NewAnalyzer(specs *Specs) *Analyzer {
	if specs == nil {
		return &Analyzer{}
	}

	return &Analyzer{
		analyzer: specs.document.Analyzer,
		schema:   specs.document.Schema(),
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

	for _, param := range operation.Parameters {
		var err error

		switch param.In {
		case "body":
			err = t.validateBodyParameter(req, &param)
		case "query":
			err = t.validateQueryParameter(req, &param)
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *Analyzer) validateBodyParameter(req *http.Request, param *spec.Parameter) error {
	bodyReader, err := req.GetBody()
	if err != nil {
		return err
	}

	input := map[string]interface{}{}
	err = json.NewDecoder(bodyReader).Decode(&input)
	if err != nil {
		return err
	}

	paramRef := param.ParamProps.Schema.Ref.String()

	var schema *spec.Schema
	for _, def := range t.analyzer.AllDefinitions() {
		if paramRef == def.Ref.String() {
			schema = def.Schema
			break
		}
	}

	err = validate.AgainstSchema(schema, input, strfmt.Default)
	if err != nil {
		return err
	}

	return nil
}

func (t *Analyzer) validateQueryParameter(req *http.Request, param *spec.Parameter) error {
	query := req.URL.Query()

	errs := validate.NewParamValidator(param, strfmt.Default).Validate(query[param.Name])
	if errs != nil {
		return errs.AsError()
	}

	return nil
}
