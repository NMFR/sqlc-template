package code_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/NMFR/sqlc-template/internal/code"
	"github.com/NMFR/sqlc-template/internal/protos/plugin"
)

func TestCodeGeneratorTemplateFuncs(t *testing.T) {
	testCases := map[string]struct {
		request  *plugin.GenerateRequest
		expected *plugin.GenerateResponse
	}{
		"test-buildin-funcs": {
			request: &plugin.GenerateRequest{
				PluginOptions: []byte(`{
					"filename": "test.file",
					"template": "` + jsonString(`
							{{ ReplaceAll "foo bar foo bar" "bar" "foo" }}
							{{ "FoO bAr" | ToLower }}
							{{ "FoO bAr" | ToUpper }}
							{{ "foo bar" | ToSnake }}
							{{ "foo bar" | ToScreamingSnake }}
							{{ "foo bar" | ToKebab }}
							{{ "foo bar" | ToScreamingKebab }}
							{{ "foo bar" | ToCamel }}
							{{ "foo bar" | ToLowerCamel }}
`) + `"
				}`),
			},
			expected: &plugin.GenerateResponse{
				Files: []*plugin.File{
					{
						Name: "test.file",
						Contents: []byte(`
							foo foo foo foo
							foo bar
							FOO BAR
							foo_bar
							FOO_BAR
							foo-bar
							FOO-BAR
							FooBar
							fooBar
`),
					},
				},
			},
		},
	}

	for testName, testCase := range testCases {
		testCase := testCase

		t.Run(testName, func(t *testing.T) {
			requestReader, err := requestToReader(testCase.request)
			assert.NoError(t, err)

			responseBuffer := &bytes.Buffer{}
			err = code.GenerateFromReader(requestReader, responseBuffer)
			assert.NoError(t, err)

			response, err := responseFromReader(responseBuffer)
			assert.NoError(t, err)

			assert.EqualExportedValues(t, testCase.expected, response)
		})
	}
}
