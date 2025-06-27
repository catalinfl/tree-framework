package render

import (
	"net/http"

	"gopkg.in/yaml.v3"
)

type YAML struct {
	Data any
}

func (y *YAML) Render(w http.ResponseWriter) error {
	if w == nil {
		return nil
	}

	y.WritingContentType(w)

	err := yaml.NewEncoder(w).Encode(y.Data)
	if err != nil {
		http.Error(w, "YAML encoding error: "+err.Error(), http.StatusInternalServerError)
	}

	return nil
}

func (y *YAML) WritingContentType(w http.ResponseWriter) error {
	writeContentType(w, []string{"application/x-yaml; charset=utf-8"})
	return nil
}
