package oaichecker

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_NewSpecsFromFile_with_load_error(t *testing.T) {
	specs, err := NewSpecsFromFile("some-unknown-path")

	assert.Nil(t, specs)
	assert.EqualError(t, err, "open some-unknown-path: no such file or directory")
}

func Test_NewSpecsFromRaw_with_unmarshalable_content(t *testing.T) {
	specs, err := NewSpecsFromRaw([]byte("no a valid spec"))

	assert.Nil(t, specs)
	assert.EqualError(t, err, "analyzed: yaml: unmarshal errors:\n  line 1: cannot unmarshal !!str `no a va...` into map[interface {}]interface {}")
}

func Test_Specs_Validate_with_invalid_specs(t *testing.T) {
	specs, err := NewSpecsFromFile("./dataset/petstore_invalid.json")
	require.NoError(t, err)

	err = specs.Validate()

	assert.EqualError(t, err, "validation failure list:\n"+
		"\"paths./pets.get.responses.200\" must validate one and only one schema (oneOf). Found none valid\n"+
		"\"paths./pets.get.responses.200.schema\" must validate one and only one schema (oneOf). Found none valid\n"+
		"items in paths./pets.get.responses.200.schema is required")
}

func Test_Specs_Validate_with_valid_specs(t *testing.T) {
	specs, err := NewSpecsFromFile("./dataset/petstore_minimal.json")
	require.NoError(t, err)

	err = specs.Validate()

	assert.NoError(t, err)
}
