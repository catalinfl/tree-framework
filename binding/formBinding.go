package binding

import (
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/url"
	"reflect"
	"strings"
)

type FormBinding struct{}

func (FormBinding) Name() string {
	return "form"
}

var defaultMemory int64 = 10 << 20 // 10 MB

func (FormBinding) Bind(req *http.Request, v any) error {
	if req == nil {
		return errors.New("request is invalid")
	}

	contentType := req.Header.Get("Content-Type")
	if contentType == "" {
		return errors.New("content type is missing")
	}

	if req.Method != http.MethodPost {
		return errors.New("method not allowed")
	}

	if strings.HasPrefix(contentType, "multipart/form-data") {
		if err := req.ParseMultipartForm(defaultMemory); err != nil {
			return errors.New("error parsing multipart form: " + err.Error())
		}
	} else {
		if err := req.ParseForm(); err != nil {
			return errors.New("error parsing form: " + err.Error())
		}
	}

	if req.PostForm == nil {
		return errors.New("post form is nil")
	}

	if err := decodeForm(req, v); err != nil {
		return errors.New("error decoding form: " + err.Error())
	}

	return validator.Validate(v)
}

func decodeForm(req *http.Request, v any) error {
	err := mapForm(req, v)
	if err != nil {
		return errors.New("error decoding form: " + err.Error())
	}

	return nil
}

func mapForm(req *http.Request, v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("binding destination must be a non-nil pointer")
	}

	elem := rv.Elem()
	if elem.Kind() != reflect.Struct {
		return errors.New("binding destination must be a pointer to a struct")
	}

	structType := elem.Type()

	for i := 0; i < elem.NumField(); i++ {
		fieldVal := elem.Field(i)
		fieldType := structType.Field(i)

		if !fieldVal.CanSet() {
			continue
		}

		formTag := fieldType.Tag.Get("form")
		fmt.Println("formTag", formTag)
		if formTag == "" {
			formTag = fieldType.Name
		}

		if fieldType.Type == reflect.TypeOf((*multipart.FileHeader)(nil)) {
			if req.MultipartForm != nil && req.MultipartForm.File != nil {
				files, found := req.MultipartForm.File[formTag]
				if found && len(files) > 0 {
					fieldVal.Set(reflect.ValueOf(files[0]))
				}
			}
			continue
		}

		if fieldType.Type == reflect.TypeOf(([]*multipart.FileHeader)(nil)) {
			if req.MultipartForm != nil && req.MultipartForm.File != nil {
				files, found := req.MultipartForm.File[formTag]
				if found {
					fieldVal.Set(reflect.ValueOf(files))
				}
			}
			continue
		}

		var params url.Values
		if req.MultipartForm != nil {
			params = req.MultipartForm.Value
		} else {
			params = req.PostForm
		}

		paramValues, found := params[formTag]
		if !found || len(paramValues) == 0 {
			continue
		}

		if err := setQueryField(fieldVal, paramValues); err != nil {
			return fmt.Errorf("failed to set field '%s' from form param '%s': %w", fieldType.Name, formTag, err)
		}
	}
	return nil
}
