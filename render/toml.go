package render

import (
	"net/http"

	"github.com/BurntSushi/toml"
)

type TOML struct {
	Data any
}

func (t *TOML) Render(w http.ResponseWriter) error {
	if w == nil {
		return nil
	}

	t.WritingContentType(w)
	if err := toml.NewEncoder(w).Encode(t.Data); err != nil {
		http.Error(w, "TOML encoding error: "+err.Error(), http.StatusInternalServerError)
	}

	return nil
}

func (t *TOML) WritingContentType(w http.ResponseWriter) error {
	writeContentType(w, []string{"application/toml; charset=utf-8"})
	return nil
}
