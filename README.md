# OpenAPI Checker

Run you server tests and at the same time check if you tests match with you OpenAPI specs.


## Why?

As a developper I often forget to change the specs after a server modification and
so my API documentation is always deprecated after 2 commits. In order to fix that
I have written a little lib which takes you tests HTTP request/responses an
matches them against your API specs. If you request doesn't matchs the specs,
you request will fail and so you test also fail.


## How?

This lib permits you to generate an http.RoundTripper which can be easily
integrated inside you http.Client from you already existing tests.


## Example

```go
import (
	"github.com/stretchr/testify/assert"
	"github.com/Peltoche/oaichecker"
)

func Test_Posting_a_valid_pet(t *testing.T) {
	// Loads the specs file.
	//
	// It contains only a the specs for the endpoint 'GET /pets'.
	specs, err := oaichecker.NewSpecsFromFile("./dataset/petstore_minimal.json")
	require.NoError(t, err)

	// Create a client which the custom transport
	client := http.Client{
		Transport: oaichecker.NewTransport(specs),
	}

	// Make the requests as usual.
	//
	// In this case the request is valid but the specs are not followed because
	// the endpoint 'POST /v2/pet' is not defined inside the specs, only 'GET /pets'.
	_, err = client.Post("http://petstore.swagger.io/v2/pet", "application/json", strings.NewReader(`{
		"name": "doggie",
		"photoUrls": ["some-url"],
	}`))

	// This assert should success but as the specs are not followed, `req` is
	// nil and `err` contains the following message:
	//
	// "Post http://petstore.swagger.io/v2/pet: operation not defined inside the specs"
	assert.NoError(t, err)
}

```
