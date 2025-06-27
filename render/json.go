package render

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type SecureJSON struct {
	Prefix string
	Data   any
}

type JSON struct {
	Data any
}

type JSONP struct {
	Callback string
	Data     any
}

type PureJSON struct {
	Data any
}

type ASCIIJSON struct {
	Data any
}

var jsonContentType = []string{"application/json; charset=utf-8"}
var jsonpContentType = []string{"application/javascript; charset=utf-8"}
var asciiJSONContentType = []string{"application/json"}

func writeJSON(w http.ResponseWriter, data any) error {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		http.Error(w, "json marshal error: "+err.Error(), http.StatusInternalServerError)
		return err
	}

	_, err = w.Write(jsonBytes)
	if err != nil {
		http.Error(w, "write json error: "+err.Error(), http.StatusInternalServerError)
		return err
	}

	return nil
}

func (j JSON) Render(w http.ResponseWriter) error {
	writeContentType(w, jsonContentType)
	return writeJSON(w, j.Data)
}

func (s SecureJSON) Render(w http.ResponseWriter) error {
	writeContentType(w, jsonContentType)
	jsonBytes, err := json.Marshal(s.Data)
	if err != nil {
		http.Error(w, "json marshal error: "+err.Error(), http.StatusInternalServerError)
		return err
	}

	if bytes.HasPrefix(jsonBytes, []byte("[")) && bytes.HasSuffix(jsonBytes, []byte("]")) {
		if _, err := w.Write([]byte(s.Prefix)); err != nil {
			http.Error(w, "write prefix error: "+err.Error(), http.StatusInternalServerError)
			return err
		}
	}

	if _, err := w.Write(jsonBytes); err != nil {
		http.Error(w, "write json error: "+err.Error(), http.StatusInternalServerError)
	}

	return nil
}

func (jp JSONP) Render(w http.ResponseWriter) error {
	if jp.Callback == "" {
		http.Error(w, "callback is empty", http.StatusBadRequest)
		return nil
	}

	writeContentType(w, jsonpContentType)

	jsonBytes, err := json.Marshal(jp.Data)
	if err != nil {
		http.Error(w, "json marshal error: "+err.Error(), http.StatusInternalServerError)
		return err
	}

	if _, err := w.Write([]byte(jp.Callback)); err != nil {
		http.Error(w, "write callback error: "+err.Error(), http.StatusInternalServerError)
		return err
	}

	if _, err := w.Write([]byte("(")); err != nil {
		http.Error(w, "write json error: "+err.Error(), http.StatusInternalServerError)
		return err
	}

	if _, err := w.Write(jsonBytes); err != nil {
		http.Error(w, "write json error: "+err.Error(), http.StatusInternalServerError)
		return err
	}

	if _, err := w.Write([]byte(");")); err != nil {
		http.Error(w, "write json error: "+err.Error(), http.StatusInternalServerError)
		return err
	}

	return nil
}

func (pj PureJSON) Render(w http.ResponseWriter) error {
	writeContentType(w, jsonContentType)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	return enc.Encode(pj.Data)
}

func (aj ASCIIJSON) Render(w http.ResponseWriter) error {
	writeContentType(w, asciiJSONContentType)

	jsonBytes, err := json.Marshal(aj.Data)
	if err != nil {
		return err
	}

	if len(jsonBytes) > 0 && jsonBytes[len(jsonBytes)-1] == '\n' {
		jsonBytes = jsonBytes[:len(jsonBytes)-1]
	}

	jsonString := string(jsonBytes)
	var result strings.Builder

	for _, r := range jsonString {
		if r <= 127 {
			result.WriteRune(r)
		} else {
			result.WriteString(fmt.Sprintf("\\u%04x", r))
		}
	}

	_, err = w.Write([]byte(result.String()))
	return err
}
