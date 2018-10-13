package oaichecker

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_NewSpecsFromFile_with_load_error(t *testing.T) {
	specs, err := NewSpecsFromFile("some-unknown-path")

	assert.Nil(t, specs)
	assert.EqualError(t, err, "open some-unknown-path: no such file or directory")
}

func Test_NewSpecsFromRaw(t *testing.T) {
	rawSpecs, err := ioutil.ReadFile("./dataset/petstore_minimal.json")
	require.NoError(t, err)

	specs, err := NewSpecsFromRaw(rawSpecs)

	assert.NotNil(t, specs)
	assert.NoError(t, err)
}

func Test_NewSpecsFromRaw_with_unmarshalable_content(t *testing.T) {
	specs, err := NewSpecsFromRaw([]byte("no a valid spec"))

	assert.Nil(t, specs)
	assert.EqualError(t, err, "analyzed: yaml: unmarshal errors:\n"+
		"  line 1: cannot unmarshal !!str `no a va...` into map[interface {}]interface {}")
}

func Test_NewSpecsFromFile_with_multi_file_spec_and_invalid_ref_should_fail(t *testing.T) {
	specs, err := NewSpecsFromFile("./dataset/petstore_invalid_ref.json")

	assert.Nil(t, specs)
	assert.Contains(t, err.Error(), "no such file or directory")
}

func Test_NewSpecsFromFile_with_multi_file_spec(t *testing.T) {
	specs, err := NewSpecsFromFile("./dataset/multi_file_spec/petstore_minimal.json")

	assert.NotNil(t, specs)
	assert.NoError(t, err)
}

func Test_NewSpecsFromRaw_with_multi_file_spec_and_invalid_working_dir_should_fail(t *testing.T) {
	rawSpecs, err := ioutil.ReadFile("./dataset/multi_file_spec/petstore_minimal.json")
	require.NoError(t, err)

	specs, err := NewSpecsFromRaw(rawSpecs)

	assert.Nil(t, specs)
	assert.Contains(t, err.Error(), "no such file or directory")
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
