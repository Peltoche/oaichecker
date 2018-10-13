package oaichecker

import (
	"encoding/json"

	"github.com/go-openapi/analysis"
	"github.com/go-openapi/loads"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"
)

type Specs struct {
	document *loads.Document
}

func NewSpecsFromFile(path string) (*Specs, error) {
	doc, err := loads.Spec(path)
	if err != nil {
		return nil, err
	}

	err = analysis.Flatten(analysis.FlattenOpts{
		Spec:     doc.Analyzer,
		BasePath: path,
		Expand:   true,
	})
	if err != nil {
		return nil, err
	}

	spec := Specs{
		document: doc,
	}

	return &spec, nil
}

func NewSpecsFromRaw(rawSpec []byte) (*Specs, error) {
	document, err := loads.Analyzed(json.RawMessage(rawSpec), "")
	if err != nil {
		return nil, err
	}

	err = analysis.Flatten(analysis.FlattenOpts{
		Spec:     document.Analyzer,
		BasePath: "",
		Expand:   true,
	})
	if err != nil {
		return nil, err
	}

	spec := Specs{
		document: document,
	}

	return &spec, nil
}

func (t *Specs) Validate() error {
	validator := validate.NewSpecValidator(t.document.Schema(), strfmt.Default)

	errs, _ := validator.Validate(t.document)
	if len(errs.Errors) > 0 {
		return errs.AsError()
	}

	return nil
}
