package binding

import (
	"errors"
	"io"
	"net/http"
	"reflect"
)

type TextBinding struct{}

func (TextBinding) Name() string {
	return "text"
}

func (TextBinding) Bind(req *http.Request, v any) error {
	if req == nil || req.Body == nil {
		return errors.New("request is invalid")
	}

	return decodeText(req.Body, v)
}

func decodeText(r io.Reader, v any) error {
	n, err := io.ReadAll(r)
	if err != nil {
		return errors.New("error decoding text: " + err.Error())
	}

	refVal := reflect.ValueOf(v)

	if refVal.Kind() != reflect.Ptr || refVal.IsNil() {
		return errors.New("error decoding text: v must be a pointer")
	}

	elem := refVal.Elem()
	if elem.Kind() == reflect.String {
		elem.SetString(BytesToString(n))
		return nil
	}

	return validator.Validate(v)
}
