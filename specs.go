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

// NewSpecsFromFile load a new OpenAPI specifications file from a filepath.
//
// This specs can be either in JSON or YAML format. Any relatives references
// ("$ref": "./pet_definition.json" for example) will be resolved based on the
// given path.
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

// NewSpecsFromFile load a new raw OpenAPI specifications.
//
// This specs can be either in JSON or YAML format. Any relatives references
// ("$ref": "./pet_definition.json" for example) will be resolved based on the
// result of os.Getwd().
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

// Validate the specs correctness.
//
// It checks that:
// - Definition can't declare a property that's already defined by one of its ancestors
// - Definition's ancestor can't be a descendant of the same model
// - Path uniqueness: each api path should be non-verbatim (account for path param names) unique per method
// - Each security reference should contain only unique scopes
// - Each security scope in a security definition should be unique
// - Parameters in path must be unique
// - Each path parameter must correspond to a parameter placeholder and vice versa
// - Each referenceable definition must have references
// - Each definition property listed in the required array must be defined in the properties of the model
// - Each parameter should have a unique `name` and `type` combination
// - Each operation should have only 1 parameter of type body
// - Each reference must point to a valid object
// - Every default value that is specified must validate against the schema for that property
// - Items property is required for all schemas/definitions of type `array`
// - Path parameters must be declared a required
// - Headers must not contain $ref
// - Schema and property examples provided must validate against their respective object's schema
// - Examples provided must validate their schema
func (t *Specs) Validate() error {
	validator := validate.NewSpecValidator(t.document.Schema(), strfmt.Default)

	errs, _ := validator.Validate(t.document)
	if len(errs.Errors) > 0 {
		return errs.AsError()
	}

	return nil
}
