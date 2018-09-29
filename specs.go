package oaichecker

import (
	"encoding/json"
	"io/ioutil"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"
)

type Specs struct {
	document *loads.Document
}

func NewSpecsFromFile(path string) (*Specs, error) {
	rawFile, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return NewSpecsFromRaw(rawFile)
}

func NewSpecsFromRaw(rawSpec []byte) (*Specs, error) {
	document, err := loads.Analyzed(json.RawMessage(rawSpec), "")
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
