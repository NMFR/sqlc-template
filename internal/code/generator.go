package code

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"text/template"

	"google.golang.org/protobuf/proto"

	"github.com/NMFR/sqlc-template/internal/protos/plugin"
)

type pluginOptions struct {
	Filename *string `json:"filename,omitempty"`
	Template *string `json:"template,omitempty"`
}

func Generate(request *plugin.GenerateRequest) (*plugin.GenerateResponse, error) {
	pluginOptions := &pluginOptions{}
	if err := json.Unmarshal(request.GetPluginOptions(), pluginOptions); err != nil {
		return nil, fmt.Errorf("failed to parse the sqlc config 'sql[].codegen.options' field to JSON, %w", err)
	}

	if pluginOptions.Filename == nil {
		return nil, fmt.Errorf("missing the sqlc config 'sql[].codegen.options.filename' field")
	}

	if pluginOptions.Template == nil {
		return nil, fmt.Errorf("missing the sqlc 'sql[].codegen.options.template' field")
	}

	tmpl, err := template.
		New("template").
		Funcs(getTemplateFunctions()).
		Parse(*pluginOptions.Template)
	if err != nil {
		return nil, fmt.Errorf("failed to parse the template, %w", err)
	}

	buf := bytes.Buffer{}
	if err := tmpl.Execute(&buf, request); err != nil {
		return nil, fmt.Errorf("failed to execute the template, %w", err)
	}

	return &plugin.GenerateResponse{
			Files: []*plugin.File{
				{Name: *pluginOptions.Filename, Contents: buf.Bytes()},
			},
		},
		nil
}

func GenerateFromBytes(in []byte) ([]byte, error) {
	request := &plugin.GenerateRequest{}
	if err := proto.Unmarshal(in, request); err != nil {
		return nil, fmt.Errorf("failed to parse / unmarshal the sqlc plugin GenerateRequest, %w", err)
	}

	response, err := Generate(request)
	if err != nil {
		return nil, err
	}

	out, err := proto.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("failed to format / marshal the sqlc plugin GenerateResponse, %w", err)
	}

	return out, nil
}

func GenerateFromReader(reader io.Reader, writer io.Writer) error {
	in, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read from stream, %w", err)
	}

	out, err := GenerateFromBytes(in)
	if err != nil {
		return err
	}

	if _, err = writer.Write(out); err != nil {
		return fmt.Errorf("failed to write to the stream, %w", err)
	}

	return nil
}
