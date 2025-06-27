package render

import (
	"encoding/xml"
	"net/http"
)

type XML struct {
	Data any
}

func (x *XML) Render(w http.ResponseWriter) error {
	if w == nil {
		return nil
	}

	x.WritingContentType(w)
	err := xml.NewEncoder(w).Encode(x.Data)
	if err != nil {
		http.Error(w, "XML encoding error: "+err.Error(), http.StatusInternalServerError)
	}

	return nil
}

func (x *XML) WritingContentType(w http.ResponseWriter) error {
	writeContentType(w, []string{"application/xml; charset=utf-8"})
	return nil
}
