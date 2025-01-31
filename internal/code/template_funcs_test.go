package code_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/NMFR/sqlc-template/internal/code"
	"github.com/NMFR/sqlc-template/internal/protos/plugin"
)

func createTemplateTestGenerateRequest(content string) *plugin.GenerateRequest {
	return &plugin.GenerateRequest{
		PluginOptions: []byte(`{
		"filename": "",
		"template": "` + jsonString(content) + `"
	}`),
	}
}

func createTemplateTestGenerateResponse(content string) *plugin.GenerateResponse {
	return &plugin.GenerateResponse{
		Files: []*plugin.File{
			{
				Name:     "",
				Contents: []byte(content),
			},
		},
	}
}

func TestCodeGeneratorTemplateFuncs(t *testing.T) {
	testCases := map[string]struct {
		request  *plugin.GenerateRequest
		expected *plugin.GenerateResponse
	}{
		"ReplaceAll": {
			request:  createTemplateTestGenerateRequest(`{{ ReplaceAll "foo bar foo bar" "bar" "foo" }}`),
			expected: createTemplateTestGenerateResponse(`foo foo foo foo`),
		},
		"ToLower": {
			request:  createTemplateTestGenerateRequest(`{{ "FoO bAr" | ToLower }}`),
			expected: createTemplateTestGenerateResponse(`foo bar`),
		},
		"ToUpper": {
			request:  createTemplateTestGenerateRequest(`{{ "FoO bAr" | ToUpper }}`),
			expected: createTemplateTestGenerateResponse(`FOO BAR`),
		},
		"ToSnake": {
			request:  createTemplateTestGenerateRequest(`{{ "foo bar" | ToSnake }}`),
			expected: createTemplateTestGenerateResponse(`foo_bar`),
		},
		"ToScreamingSnake": {
			request:  createTemplateTestGenerateRequest(`{{ "foo bar" | ToScreamingSnake }}`),
			expected: createTemplateTestGenerateResponse(`FOO_BAR`),
		},
		"ToKebab": {
			request:  createTemplateTestGenerateRequest(`{{ "foo bar" | ToKebab }}`),
			expected: createTemplateTestGenerateResponse(`foo-bar`),
		},
		"ToScreamingKebab": {
			request:  createTemplateTestGenerateRequest(`{{ "foo bar" | ToScreamingKebab }}`),
			expected: createTemplateTestGenerateResponse(`FOO-BAR`),
		},
		"ToCamel": {
			request:  createTemplateTestGenerateRequest(`{{ "foo bar" | ToCamel }}`),
			expected: createTemplateTestGenerateResponse(`FooBar`),
		},
		"ToLowerCamel": {
			request:  createTemplateTestGenerateRequest(`{{ "foo bar" | ToLowerCamel }}`),
			expected: createTemplateTestGenerateResponse(`fooBar`),
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

			assert.Equal(t, string(testCase.expected.Files[0].Contents), string(response.Files[0].Contents))
		})
	}
}
