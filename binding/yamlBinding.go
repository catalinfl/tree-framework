package binding

import (
	"errors"
	"io"
	"net/http"

	"gopkg.in/yaml.v3"
)

type YAMLBinding struct{}

func (YAMLBinding) Name() string {
	return "html"
}

func (YAMLBinding) Bind(req *http.Request, v any) error {
	if req == nil || req.Body == nil {
		return errors.New("request is invalid")
	}

	return decodeYAML(req.Body, v)
}

func decodeYAML(r io.Reader, v any) error {
	decoder := yaml.NewDecoder(r)

	if err := decoder.Decode(v); err != nil {
		return errors.New("error decoding yaml: " + err.Error())
	}
	return nil
}
