package binding

import "net/http"

type Binding interface {
	Name() string
	Bind(*http.Request, any) error
}

type URIBindingInterface interface {
	Name() string
	BindURI(map[string]string, any) error
}

type QueryBindingInterface interface {
	Name() string
	BindQuery(*http.Request, any) error
}

var (
	JSON  Binding               = &JSONBinding{}
	XML   Binding               = &XMLBinding{}
	YAML  Binding               = &YAMLBinding{}
	Text  Binding               = &TextBinding{}
	TOML  Binding               = &TOMLBinding{}
	Form  Binding               = &FormBinding{}
	URI   URIBindingInterface   = &URIBinding{}
	Query QueryBindingInterface = &QueryBinding{}
)

func Default(method, contentType string) string {
	if method == http.MethodGet {
		return "form"
	}
	return "test"
}
