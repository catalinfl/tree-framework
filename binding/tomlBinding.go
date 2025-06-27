package binding

import (
	"errors"
	"io"
	"net/http"

	"github.com/BurntSushi/toml"
)

type TOMLBinding struct{}

func (TOMLBinding) Name() string {
	return "toml"
}

func (TOMLBinding) Bind(req *http.Request, v any) error {
	if req == nil || req.Body == nil {
		return errors.New("request is invalid")
	}

	return decodeTOML(req.Body, v)
}

func decodeTOML(r io.Reader, v any) error {

	decoder := toml.NewDecoder(r)
	if _, err := decoder.Decode(v); err != nil {
		return errors.New("error decoding xml: " + err.Error())
	}
	validator.Validate(v)
	return nil
}
