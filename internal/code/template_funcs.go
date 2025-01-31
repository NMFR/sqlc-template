package code

import (
	"strings"
	"text/template"

	"github.com/iancoleman/strcase"
)

func getTemplateFunctions() template.FuncMap {
	return template.FuncMap{
		"ReplaceAll": strings.ReplaceAll,
		"ToLower":    strings.ToLower,
		"ToUpper":    strings.ToUpper,

		"ToSnake":          strcase.ToSnake,          // ("foo bar") =>	"foo_bar"
		"ToScreamingSnake": strcase.ToScreamingSnake, // ("foo bar") => "FOO_BAR"
		"ToKebab":          strcase.ToKebab,          // ("foo bar") =>	"foo-bar"
		"ToScreamingKebab": strcase.ToScreamingKebab, // ("foo bar") =>	"FOO-BAR"
		"ToCamel":          strcase.ToCamel,          // ("foo bar") =>	"FooBar"
		"ToLowerCamel":     strcase.ToLowerCamel,     // ("foo bar") =>	"fooBar"
	}
}
