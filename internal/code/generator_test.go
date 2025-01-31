package code_test

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"

	"github.com/NMFR/sqlc-template/internal/code"
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

func requestToReader(request *plugin.GenerateRequest) (io.Reader, error) {
	requestBytes, err := proto.Marshal(request)
	if err != nil {
		return nil, err
	}

	return bytes.NewBuffer(requestBytes), nil
}

func requestToReaderNoErr(request *plugin.GenerateRequest) io.Reader {
	response, err := requestToReader(request)
	if err != nil {
		panic(err)
	}

	return response
}

func responseFromReader(reader io.Reader) (*plugin.GenerateResponse, error) {
	responseBytes, err := io.ReadAll(reader)
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

type errorReaderWriter struct{}

func (rw *errorReaderWriter) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("read error")
}

func (rw *errorReaderWriter) Write(p []byte) (n int, err error) {
	return 0, fmt.Errorf("write error")
}

func TestCodeGeneratorSuccess(t *testing.T) {
	testCases := map[string]struct {
		request  *plugin.GenerateRequest
		expected *plugin.GenerateResponse
	}{
		"empty": {
			request: &plugin.GenerateRequest{
				PluginOptions: []byte(`{
					"filename": "test.file",
					"template": ""
				}`),
			},
			expected: &plugin.GenerateResponse{
				Files: []*plugin.File{
					{
						Name:     "test.file",
						Contents: nil,
					},
				},
			},
		},
		"simple-template": {
			request: &plugin.GenerateRequest{
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
			expected: &plugin.GenerateResponse{
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

func TestCodeGeneratorFailure(t *testing.T) {
	testCases := map[string]struct {
		requestReader  io.Reader
		response       io.Writer
		expectedErrMsg string
	}{
		"bad reader": {
			requestReader:  &errorReaderWriter{},
			response:       &bytes.Buffer{},
			expectedErrMsg: "failed to read from stream",
		},
		"bad writer": {
			requestReader: requestToReaderNoErr(&plugin.GenerateRequest{
				PluginOptions: []byte(`{
					"filename": "test.file",
					"template": ""
				}`),
			}),
			response:       &errorReaderWriter{},
			expectedErrMsg: "failed to write to the stream",
		},
		"bad reader data": {
			requestReader:  bytes.NewBuffer([]byte("not a binary protobuf message at all")),
			response:       &bytes.Buffer{},
			expectedErrMsg: "failed to parse / unmarshal the sqlc plugin GenerateRequest",
		},
		"empty options": {
			requestReader:  requestToReaderNoErr(&plugin.GenerateRequest{}),
			response:       &bytes.Buffer{},
			expectedErrMsg: "failed to parse the sqlc config 'sql[].codegen.options' field to JSON",
		},
		"invalid options JSON": {
			requestReader: requestToReaderNoErr(&plugin.GenerateRequest{
				PluginOptions: []byte("not a JSON at all"),
			}),
			response:       &bytes.Buffer{},
			expectedErrMsg: "failed to parse the sqlc config 'sql[].codegen.options' field to JSON",
		},
		"empty options JSON": {
			requestReader: requestToReaderNoErr(&plugin.GenerateRequest{
				PluginOptions: []byte("{}"),
			}),
			response:       &bytes.Buffer{},
			expectedErrMsg: "missing the sqlc config 'sql[].codegen.options.filename' field",
		},
		"missing template option": {
			requestReader: requestToReaderNoErr(&plugin.GenerateRequest{
				PluginOptions: []byte(`{
					"filename": "test.file"
				}`),
			}),
			response:       &bytes.Buffer{},
			expectedErrMsg: "missing the sqlc 'sql[].codegen.options.template' field",
		},
		"missing filename option": {
			requestReader: requestToReaderNoErr(&plugin.GenerateRequest{
				PluginOptions: []byte(`{
					"template": "nothing"
				}`),
			}),
			response:       &bytes.Buffer{},
			expectedErrMsg: "missing the sqlc config 'sql[].codegen.options.filename' field",
		},
		"invalid template option": {
			requestReader: requestToReaderNoErr(&plugin.GenerateRequest{
				PluginOptions: []byte(`{
					"filename": "test.file",
					"template": "{{ invalid"
				}`),
			}),
			response:       &bytes.Buffer{},
			expectedErrMsg: "failed to parse the template",
		},
		"template option using non existant data": {
			requestReader: requestToReaderNoErr(&plugin.GenerateRequest{
				PluginOptions: []byte(`{
					"filename": "test.file",
					"template": "{{ .invalid }}"
				}`),
			}),
			response:       &bytes.Buffer{},
			expectedErrMsg: "failed to execute the template",
		},
	}

	for testName, testCase := range testCases {
		testCase := testCase

		t.Run(testName, func(t *testing.T) {
			err := code.GenerateFromReader(testCase.requestReader, testCase.response)

			assert.Error(t, err)
			assert.ErrorContains(t, err, testCase.expectedErrMsg)
		})
	}
}
