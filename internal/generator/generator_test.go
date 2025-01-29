package generator_test

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"

	"github.com/NMFR/sqlc-template/internal/generator"
	"github.com/NMFR/sqlc-template/internal/protos/plugin"
)

// jsonString escapes a string for a valid string property in a json document.
func jsonString(str string) string {
	for old, new := range map[string]string{
		"\n": "\\n",
		"\t": "\\t",
		"\"": "\\\"",
	} {
		str = strings.ReplaceAll(str, old, new)
	}

	return str
}

// runGenerateFromReader run the `generator.GenerateFromReader` by marshaling the `plugin.GenerateRequest` input to bytes and unmarshaling the result back to a `plugin.GenerateResponse`.
func runGenerateFromReader(request *plugin.GenerateRequest) (*plugin.GenerateResponse, error) {
	requestBytes, err := proto.Marshal(request)
	if err != nil {
		return nil, err
	}

	requestReader := bytes.NewBuffer(requestBytes)
	responseWriter := bytes.Buffer{}

	err = generator.GenerateFromReader(requestReader, &responseWriter)
	if err != nil {
		return nil, err
	}

	responseBytes, err := io.ReadAll(&responseWriter)
	if err != nil {
		return nil, err
	}

	response := &plugin.GenerateResponse{}
	err = proto.Unmarshal(responseBytes, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func TestGeneratorSuccess(t *testing.T) {
	testCases := map[string]struct {
		request  plugin.GenerateRequest
		expected plugin.GenerateResponse
	}{
		"empty": {
			request: plugin.GenerateRequest{
				PluginOptions: []byte(`{
					"filename": "test.file",
					"template": ""
				}`),
			},
			expected: plugin.GenerateResponse{
				Files: []*plugin.File{
					{
						Name:     "test.file",
						Contents: nil,
					},
				},
			},
		},
		"test-buildin-funcs": {
			request: plugin.GenerateRequest{
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
			expected: plugin.GenerateResponse{
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
		"simple-template": {
			request: plugin.GenerateRequest{
				Queries: []*plugin.Query{
					{
						Name: "mock query",
						Cmd:  ":one",
						Params: []*plugin.Parameter{
							{
								Column: &plugin.Column{
									Name: "mock param",
									Type: &plugin.Identifier{Name: "text"},
								},
							},
							{
								Column: &plugin.Column{
									Name: "another mock param",
									Type: &plugin.Identifier{Name: "int"},
								},
							},
						},
						Columns: []*plugin.Column{
							{
								Name: "mock column",
								Type: &plugin.Identifier{Name: "int"},
							},
							{
								Name: "another mock column",
								Type: &plugin.Identifier{Name: "bool"},
							},
						},
					},
				},
				PluginOptions: []byte(`{
					"filename": "some.file",
					"template": "` + jsonString(`
							queries:
							{{- range .Queries }}
								- name: {{ .Name | ToLowerCamel }}
									cmd: {{ .Cmd }}
									params:
									{{- range .Params }}
										- name: {{ .Column.Name | ToLowerCamel }}
											type: {{ .Column.Type.Name -}}
									{{ end }}
									columns:
									{{- range .Columns }}
										- name: {{ .Name | ToLowerCamel }}
											type: {{ .Type.Name -}}
									{{ end -}}
							{{ end }}
`) + `"
				}`),
			},
			expected: plugin.GenerateResponse{
				Files: []*plugin.File{
					{
						Name: "some.file",
						Contents: []byte(`
							queries:
								- name: mockQuery
									cmd: :one
									params:
										- name: mockParam
											type: text
										- name: anotherMockParam
											type: int
									columns:
										- name: mockColumn
											type: int
										- name: anotherMockColumn
											type: bool
`),
					},
				},
			},
		},
	}

	for testName, testCase := range testCases {
		testCase := testCase

		t.Run(testName, func(t *testing.T) {
			response, err := runGenerateFromReader(&testCase.request)

			assert.NoError(t, err)
			assert.EqualExportedValues(t, &testCase.expected, response)
		})
	}
}

func TestGeneratorFailure(t *testing.T) {
	testCases := map[string]struct {
		request        plugin.GenerateRequest
		expectedErrMsg string
	}{
		"empty options": {
			request:        plugin.GenerateRequest{},
			expectedErrMsg: "failed to parse the sqlc config 'sql[].codegen.options' field to JSON",
		},
		"invalid options JSON": {
			request: plugin.GenerateRequest{
				PluginOptions: []byte("not a JSON at all"),
			},
			expectedErrMsg: "failed to parse the sqlc config 'sql[].codegen.options' field to JSON",
		},
		"empty options JSON": {
			request: plugin.GenerateRequest{
				PluginOptions: []byte("{}"),
			},
			expectedErrMsg: "missing the sqlc config 'sql[].codegen.options.filename' field",
		},
		"missing template option": {
			request: plugin.GenerateRequest{
				PluginOptions: []byte(`{
					"filename": "test.file"
				}`),
			},
			expectedErrMsg: "missing the sqlc 'sql[].codegen.options.template' field",
		},
		"missing filename option": {
			request: plugin.GenerateRequest{
				PluginOptions: []byte(`{
					"template": "nothing"
				}`),
			},
			expectedErrMsg: "missing the sqlc config 'sql[].codegen.options.filename' field",
		},
		"invalid template option": {
			request: plugin.GenerateRequest{
				PluginOptions: []byte(`{
					"filename": "test.file",
					"template": "{{ invalid"
				}`),
			},
			expectedErrMsg: "failed to parse the template",
		},
		"template option using non existant data": {
			request: plugin.GenerateRequest{
				PluginOptions: []byte(`{
					"filename": "test.file",
					"template": "{{ .invalid }}"
				}`),
			},
			expectedErrMsg: "failed to execute the template",
		},
	}

	for testName, testCase := range testCases {
		testCase := testCase

		t.Run(testName, func(t *testing.T) {
			response, err := runGenerateFromReader(&testCase.request)

			assert.Error(t, err)
			assert.Nil(t, response)
			assert.ErrorContains(t, err, testCase.expectedErrMsg)
		})
	}
}
