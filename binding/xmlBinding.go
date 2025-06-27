package binding

import (
	"encoding/xml"
	"errors"
	"io"
	"net/http"
)

type XMLBinding struct{}

func (XMLBinding) Name() string {
	return "xml"
}

func (XMLBinding) Bind(req *http.Request, v any) error {
	if req == nil || req.Body == nil {
		return errors.New("request is invalid")
	}

	return decodeXML(req.Body, v)
}

func decodeXML(r io.Reader, v any) error {
	decoder := xml.NewDecoder(r)
	if err := decoder.Decode(v); err != nil {
		return errors.New("error decoding xml: " + err.Error())
	}
	validator.Validate(v)
	return nil
}
