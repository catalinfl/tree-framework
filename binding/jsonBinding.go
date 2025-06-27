package binding

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type JSONBinding struct{}

func (JSONBinding) Name() string {
	return "json"
}

func (JSONBinding) Bind(req *http.Request, v any) error {
	if req == nil || req.Body == nil {
		fmt.Println("request is nil")
		return errors.New("request is invalid")
	}

	return decodeJSON(req.Body, v)
}

func (JSONBinding) BindBody(data []byte, v any) any {
	return decodeJSON(bytes.NewReader(data), v)
}

// this should return error not any, must validate
func decodeJSON(r io.Reader, v any) error {
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(v); err != nil {
		return errors.New("error decoding json: " + err.Error())
	}
	return validator.Validate(v)
}
