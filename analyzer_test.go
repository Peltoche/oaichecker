package oaichecker

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_NewAnalyzer_Analyze_with_specs(t *testing.T) {
	assert.Panics(t, func() {
		_ = NewAnalyzer(nil)
	})
}

func Test_Analyzer_Analyze_with_no_request(t *testing.T) {
	specs, err := NewSpecsFromFile("./dataset/petstore_minimal.json")
	require.NoError(t, err)

	analyzer := NewAnalyzer(specs)

	err = analyzer.Analyze(nil)

	assert.EqualError(t, err, "no request defined")
}

func Test_Analyzer_Analyze_with_request_not_in_specs(t *testing.T) {
	specs, err := NewSpecsFromFile("./dataset/petstore_minimal.json")
	require.NoError(t, err)

	analyzer := NewAnalyzer(specs)

	req, err := http.NewRequest("GET", "invalid/path", nil)
	require.NoError(t, err)

	err = analyzer.Analyze(req)

	assert.EqualError(t, err, "operation not defined inside the specs")
}

func Test_Analyzer_Analyze_with_body_parameters(t *testing.T) {
	specs, err := NewSpecsFromFile("./dataset/petstore.json")
	require.NoError(t, err)

	analyzer := NewAnalyzer(specs)

	req, err := http.NewRequest("POST", "/pet", strings.NewReader(`{
		"name": "foobar",
		"photoUrls": ["tutu"]
	}`))
	require.NoError(t, err)

	err = analyzer.Analyze(req)

	assert.NoError(t, err)
}

func Test_Analyzer_Analyze_with_invalid_body_parameters(t *testing.T) {
	specs, err := NewSpecsFromFile("./dataset/petstore.json")
	require.NoError(t, err)

	analyzer := NewAnalyzer(specs)

	req, err := http.NewRequest("POST", "/pet", strings.NewReader(`{
		"name": "foobar"
	}`))
	require.NoError(t, err)

	err = analyzer.Analyze(req)

	assert.EqualError(t, err, "validation failure list:\n"+
		".photoUrls in body is required")
}

func Test_Analyzer_Analyze_with_invalid_body_format(t *testing.T) {
	specs, err := NewSpecsFromFile("./dataset/petstore.json")
	require.NoError(t, err)

	analyzer := NewAnalyzer(specs)

	req, err := http.NewRequest("POST", "/pet", strings.NewReader(`not a json`))
	require.NoError(t, err)

	err = analyzer.Analyze(req)

	assert.EqualError(t, err, "invalid character 'o' in literal null (expecting 'u')")
}

func Test_Analyzer_Analyze_with_query_parameters(t *testing.T) {
	specs, err := NewSpecsFromFile("./dataset/petstore.json")
	require.NoError(t, err)

	analyzer := NewAnalyzer(specs)

	req, err := http.NewRequest("GET", "/pet/findByStatus", nil)
	require.NoError(t, err)

	q := req.URL.Query()
	q.Set("status", "available")
	req.URL.RawQuery = q.Encode()

	err = analyzer.Analyze(req)

	assert.NoError(t, err)
}

func Test_Analyzer_Analyze_with_invalid_query_parameters(t *testing.T) {
	specs, err := NewSpecsFromFile("./dataset/petstore.json")
	require.NoError(t, err)

	analyzer := NewAnalyzer(specs)

	req, err := http.NewRequest("GET", "/pet/findByStatus", nil)
	require.NoError(t, err)

	q := req.URL.Query()
	q.Set("status", "invalid-enum-value")
	req.URL.RawQuery = q.Encode()

	err = analyzer.Analyze(req)

	assert.EqualError(t, err, "validation failure list:\n"+
		"status.0 in query should be one of [available pending sold]")
}

func Test_Analyzer_Analyze_with_path_parameters(t *testing.T) {
	specs, err := NewSpecsFromFile("./dataset/petstore.json")
	require.NoError(t, err)

	analyzer := NewAnalyzer(specs)

	req, err := http.NewRequest("GET", "/pet/42", nil)
	require.NoError(t, err)

	err = analyzer.Analyze(req)

	assert.NoError(t, err)
}

func Test_Analyzer_Analyze_with_invalid_path_parameters(t *testing.T) {
	specs, err := NewSpecsFromFile("./dataset/petstore.json")
	require.NoError(t, err)

	analyzer := NewAnalyzer(specs)

	req, err := http.NewRequest("GET", "/pet/not-a-number", nil)
	require.NoError(t, err)

	err = analyzer.Analyze(req)

	assert.EqualError(t, err, "validation failure list:\n"+
		"petId in path must be of type integer: \"string\"")
}

func Test_Analyzer_Analyze_with_formData_file(t *testing.T) {
	specs, err := NewSpecsFromFile("./dataset/petstore.json")
	require.NoError(t, err)

	analyzer := NewAnalyzer(specs)

	var buf bytes.Buffer
	mp := multipart.NewWriter(&buf)
	fileWriter, err := mp.CreateFormFile("file", "file")
	require.NoError(t, err)
	_, err = fileWriter.Write([]byte("some-data"))
	require.NoError(t, err)

	err = mp.Close()
	require.NoError(t, err)

	req, err := http.NewRequest("POST", "/pet/32/uploadImage", &buf)
	require.NoError(t, err)

	req.Header.Set("Content-Type", mp.FormDataContentType())

	err = analyzer.Analyze(req)

	assert.NoError(t, err)
}
