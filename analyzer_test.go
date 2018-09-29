package oaichecker

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Analyzer_Analyze_with_no_specs(t *testing.T) {
	analyzer := NewAnalyzer(nil)

	err := analyzer.Analyze(nil)

	assert.EqualError(t, err, "no specs defined")
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
