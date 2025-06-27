package render

import "net/http"

type Render interface {
	Render(http.ResponseWriter) error
	WritingContentType(w http.ResponseWriter) error
}

type JSONRender interface {
	Render(http.ResponseWriter) error
}

var (
	_ Render     = (*HTMLProduction)(nil)
	_ Render     = (*HTMLDevelopment)(nil)
	_ JSONRender = (*JSON)(nil)
	_ JSONRender = (*SecureJSON)(nil)
	_ JSONRender = (*PureJSON)(nil)
	_ JSONRender = (*JSONP)(nil)
	_ JSONRender = (*ASCIIJSON)(nil)
	_ Render     = (*TOML)(nil)
	_ Render     = (*YAML)(nil)
	_ Render     = (*XML)(nil)
)

func writeContentType(w http.ResponseWriter, contentType []string) {
	header := w.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = contentType
	}
}
