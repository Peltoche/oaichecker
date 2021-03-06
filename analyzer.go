package oaichecker

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-openapi/analysis"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware/denco"
	"github.com/go-openapi/spec"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"
)

// Analyzer analyze a pair of http.Request/http.Response with the previously
// Specs loaded at initialization.
type Analyzer struct {
	analyzer *analysis.Spec
	schema   *spec.Schema
	router   *denco.Router
}

// NewAnalyzer instantiate a new Analyzer based on the given Specs.
//
// If the specs is nil, the function panics.
func NewAnalyzer(specs *Specs) *Analyzer {
	if specs == nil {
		panic("specs is nil")
	}

	return &Analyzer{
		analyzer: specs.document.Analyzer,
		schema:   specs.document.Schema(),
		router:   createRouter(specs.document.Analyzer),
	}
}

func createRouter(analyzer *analysis.Spec) *denco.Router {
	var records []denco.Record
	for _, paths := range analyzer.Operations() {
		for pathName := range paths {
			// Go from the OAI path definition to the denco format:
			//
			// i.e : "/foo/{variable}/bar" => "/foo/:variable/bar"
			dencoPath := strings.Replace(pathName, "{", ":", -1)
			dencoPath = strings.Replace(dencoPath, "}", "", -1)

			records = append(records, denco.Record{
				Key:   dencoPath,
				Value: pathName,
			})
		}
	}

	r := denco.New()
	err := r.Build(records)
	if err != nil {
		panic(err)
	}

	return r
}

// Analyze the given pair of http.Request/http.Response with the previously
// loaded Specs.
//
// This method checks:
// - If the Operation exists (method / path)
// - The Parameters defined inside the Operation (path / header / body / query / formData)
// - The Response (status / body)
//
// In case of incorrectess, an error is returned.
func (t *Analyzer) Analyze(req *http.Request, res *http.Response) error {
	if req == nil {
		return errors.New("no request defined")
	}

	pathName, pathParams, ok := t.router.Lookup(req.URL.Path)
	if !ok {
		return errors.New("operation not defined inside the specs")
	}

	operation, ok := t.analyzer.OperationFor(req.Method, pathName.(string))
	if !ok {
		return errors.New("operation not defined inside the specs")
	}

	for _, param := range operation.Parameters {
		var err error

		switch param.In {
		case "path":
			err = t.validatePathParameter(pathParams, &param)
		case "header":
			err = t.validateHeaderParameter(req, &param)
		case "body":
			err = t.validateBodyParameter(req, &param)
		case "query":
			err = t.validateQueryParameter(req, &param)
		case "formData":
			err = t.validateFormDataParameter(req, &param)
		}
		if err != nil {
			return err
		}
	}

	err := t.validateResponse(res, operation.Responses)

	return err
}

func (t *Analyzer) validateResponse(res *http.Response, resSpec *spec.Responses) error {
	for status, response := range resSpec.StatusCodeResponses {
		if status == res.StatusCode {

			body, err := ioutil.ReadAll(res.Body)
			if err != nil {
				return err
			}

			res.Body = ioutil.NopCloser(bytes.NewReader(body))

			if response.ResponseProps.Schema == nil {
				if len(body) > 0 {
					return fmt.Errorf("validation failure list:\nno response body defined inside the specs but have %q", body)
				}
				return nil
			}

			var input interface{}
			err = json.Unmarshal(body, &input)
			if err != nil {
				return fmt.Errorf("validation failure list:\nfailed to parse response body: %s", err)
			}

			err = validate.AgainstSchema(response.Schema, input, strfmt.Default)
			if err != nil {
				return err
			}

			return nil
		}
	}

	return fmt.Errorf("validation failure list:\nresponse status %s not defined inside the specs", res.Status)
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

	err = validate.AgainstSchema(param.Schema, input, strfmt.Default)
	if err != nil {
		return err
	}

	return nil
}

func (t *Analyzer) validateHeaderParameter(req *http.Request, param *spec.Parameter) error {
	errs := validate.NewParamValidator(param, strfmt.Default).Validate(req.Header.Get(param.Name))
	if errs != nil {
		return errs.AsError()
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

func (t *Analyzer) validatePathParameter(pathParams denco.Params, param *spec.Parameter) error {
	var res interface{} = pathParams.Get(param.Name)
	if param.Type == "integer" {
		nParam, err := strconv.Atoi(pathParams.Get(param.Name))
		if err == nil {
			res = nParam
		}
	}

	errs := validate.NewParamValidator(param, strfmt.Default).Validate(res)
	if errs != nil {
		return errs.AsError()
	}

	return nil
}

func (t *Analyzer) validateFormDataParameter(req *http.Request, param *spec.Parameter) error {

	var res interface{}

	if param.Type == "file" {
		data, header, err := req.FormFile(param.Name)
		if err != nil && param.ParamProps.Required {
			return fmt.Errorf("validation failure list:\n%s in formData is required", param.Name)
		}

		res = runtime.File{
			Data:   data,
			Header: header,
		}
	} else {
		res = req.PostFormValue(param.Name)
	}

	errs := validate.NewParamValidator(param, strfmt.Default).Validate(res)
	if errs != nil {
		return errs.AsError()
	}

	return nil
}
