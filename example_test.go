package oaichecker_test

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Peltoche/oaichecker"
)

func Example_specs_format_validations() {
	// Load the specs file.
	specs, err := oaichecker.NewSpecsFromFile("./dataset/petstore_invalid.json")
	if err != nil {
		panic(err)
	}

	// Validate the specs file format.
	err = specs.Validate()
	if err != nil {
		fmt.Println(err)
	}

	// Output:
	// validation failure list:
	//"paths./pets.get.responses.200" must validate one and only one schema (oneOf). Found none valid
	//"paths./pets.get.responses.200.schema" must validate one and only one schema (oneOf). Found none valid
	//items in paths./pets.get.responses.200.schema is required
}

func Example_post_request_validation() {
	// Load the specs file.
	specs, err := oaichecker.NewSpecsFromFile("./dataset/petstore.json")
	if err != nil {
		panic(err)
	}

	// Create a client which matches the requests/responses against the given
	// specs.
	client := http.Client{
		Transport: oaichecker.NewTransport(specs),
	}

	// Make a request with a required field missing.
	_, err = client.Post("http://petstore.swagger.io/pet", "application/json", strings.NewReader(`{
		"photoUrls": ["some-url"]
	}`))
	if err != nil {
		fmt.Println(err)
	}

	// Output:
	// Post http://petstore.swagger.io/pet: validation failure list:
	// .name in body is required
}
