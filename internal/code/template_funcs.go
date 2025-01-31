package code

import (
	"text/template"

	"github.com/Masterminds/sprig/v3"
)

type StringTransformer = func(string) string

func getTemplateFunctions() template.FuncMap {
	funcMap := sprig.FuncMap()

	delete(funcMap, "osBase")
	delete(funcMap, "osDir")
	delete(funcMap, "osClean")
	delete(funcMap, "osExt")

	delete(funcMap, "env")
	delete(funcMap, "expandenv")

	delete(funcMap, "kindOf")
	delete(funcMap, "kindIs")
	delete(funcMap, "typeOf")
	delete(funcMap, "typeIs")
	delete(funcMap, "typeIsLike")
	delete(funcMap, "deepEqual")

	delete(funcMap, "getHostByName")

	// In order to not break compatability still define the older functions that have equivalents in the added sprig functions:
	replace := funcMap["replace"].(func(string, string, string) string)
	lower := funcMap["lower"].(StringTransformer)
	upper := funcMap["upper"].(StringTransformer)
	snakecase := funcMap["snakecase"].(StringTransformer)
	kebabcase := funcMap["kebabcase"].(StringTransformer)
	camelcase := funcMap["camelcase"].(StringTransformer)
	untitle := funcMap["untitle"].(StringTransformer)

	funcMap["ReplaceAll"] = func(str string, old string, new string) string { return replace(old, new, str) }
	funcMap["ToLower"] = lower
	funcMap["ToUpper"] = upper
	funcMap["ToSnake"] = snakecase
	funcMap["ToScreamingSnake"] = func(s string) string { return upper(snakecase(s)) }
	funcMap["ToKebab"] = kebabcase
	funcMap["ToScreamingKebab"] = func(s string) string { return upper(kebabcase(s)) }
	funcMap["ToCamel"] = camelcase
	funcMap["ToLowerCamel"] = func(s string) string { return untitle(camelcase(s)) }

	return funcMap
}
