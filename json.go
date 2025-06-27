package tree

import (
	"encoding/json"
	"net/http"
	"reflect"
)

type J map[string]any

func (j J) encodeJSON(w http.ResponseWriter) error {
	if reflect.TypeOf(j).Kind() != reflect.Map {
		return ErrInvalidJSONType
	}

	jsonData, err := json.Marshal(j)

	if err != nil {
		return ErrInvalidJSONType
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)

	return nil
}
