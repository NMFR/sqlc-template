package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"

	"github.com/iancoleman/strcase"
	"google.golang.org/protobuf/proto"

	"github.com/NMFR/sqlc-template/pkg/sqlc/plugin"
)

type PluginOptions struct {
	Filename *string `json:"filename,omitempty"`
	Template *string `json:"template,omitempty"`
}

func generate(request *plugin.GenerateRequest) (*plugin.GenerateResponse, error) {
	pluginOptions := &PluginOptions{}
	if err := json.Unmarshal(request.GetPluginOptions(), pluginOptions); err != nil {
		return nil, fmt.Errorf("failed to parse the sqlc 'plugin options' field to JSON, %w", err)
	}

	if pluginOptions.Filename == nil {
		return nil, fmt.Errorf("missing the 'filename' plugin option")
	}

	if pluginOptions.Template == nil {
		return nil, fmt.Errorf("missing the 'template' plugin option")
	}

	funcMap := template.FuncMap{
		"Replace":    strings.Replace,
		"ReplaceAll": strings.ReplaceAll,
		"ToLower":    strings.ToLower,
		"ToUpper":    strings.ToUpper,
		"TrimSpace":  strings.TrimSpace,

		"ToSnake":              strcase.ToSnake,              //(s)	any_kind_of_string
		"ToSnakeWithIgnore":    strcase.ToSnakeWithIgnore,    //(s, '.')	any_kind.of_string
		"ToScreamingSnake":     strcase.ToScreamingSnake,     //(s)	ANY_KIND_OF_STRING
		"ToKebab":              strcase.ToKebab,              //(s)	any-kind-of-string
		"ToScreamingKebab":     strcase.ToScreamingKebab,     //(s)	ANY-KIND-OF-STRING
		"ToDelimited":          strcase.ToDelimited,          //(s, '.')	any.kind.of.string
		"ToScreamingDelimited": strcase.ToScreamingDelimited, //(s, '.', ' ', true)	ANY.KIND OF.STRING
		"ToCamel":              strcase.ToCamel,              //(s)	AnyKindOfString
		"ToLowerCamel":         strcase.ToLowerCamel,         //(s)	anyKindOfString
	}

	tmpl, err := template.New("template").Funcs(funcMap).Parse(*pluginOptions.Template)
	if err != nil {
		return nil, fmt.Errorf("failed to parse the provided template, %w", err)
	}

	buf := bytes.Buffer{}
	if err := tmpl.Execute(&buf, request); err != nil {
		return nil, fmt.Errorf("failed to execute the provided template, %w", err)
	}

	json, err := json.Marshal(request)
	if err != nil {
		panic(fmt.Errorf("failed to convert to JSON, %w", err))
	}

	return &plugin.GenerateResponse{
		Files: []*plugin.File{
			{Name: *pluginOptions.Filename, Contents: buf.Bytes()},
			{Name: "input.json", Contents: []byte(json)},
		},
	}, nil
}

func main() {
	stdin, err := io.ReadAll(os.Stdin)
	if err != nil {
		panic(fmt.Errorf("failed to read from stdin, %w", err))
	}

	request := &plugin.GenerateRequest{}
	if err := proto.Unmarshal(stdin, request); err != nil {
		panic(fmt.Errorf("failed to parse / unmarshal the sqlc request, %w", err))
	}

	response, err := generate(request)
	if err != nil {
		panic(fmt.Errorf("failed to generate, %w", err))
	}

	out, err := proto.Marshal(response)
	if err != nil {
		panic(fmt.Errorf("failed to format / marshal the sqlc response, %w", err))
	}

	if _, err = os.Stdout.Write(out); err != nil {
		panic(fmt.Errorf("failed to write to stdout, %w", err))
	}
}
