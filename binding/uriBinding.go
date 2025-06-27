package binding

import (
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"reflect"
	"strings"
	"time"
)

type URIBinding struct{}

func (URIBinding) Name() string {
	return "uri"
}

func (URIBinding) BindURI(params map[string]string, v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("v must be a pointer")
	}

	elem := rv.Elem()
	if elem.Kind() != reflect.Struct {
		return errors.New("v must be a struct pointer")
	}

	if err := mapUriParams(params, elem); err != nil {
		return errors.New("error binding uri: " + err.Error())
	}

	return validator.Validate(v)
}

func mapUriParams(params map[string]string, structVal reflect.Value) error {
	structType := structVal.Type()

	for i := 0; i < structType.NumField(); i++ {
		fieldVal := structVal.Field(i)
		fieldType := structType.Field(i)

		if !fieldVal.CanSet() {
			continue
		}

		uriTag := fieldType.Tag.Get("uri")
		if uriTag == "" {
			fmt.Println("[WARNING] URI tag not found for field:", fieldType.Name)
			continue
		}

		paramValue, found := findParamCaseInsensitive(params, uriTag)
		if !found {
			continue
		}

		if err := setFieldValue(fieldVal, paramValue); err != nil {
			return fmt.Errorf("error setting field %s: %w", fieldType.Name, err)
		}
	}

	return nil
}

func findParamCaseInsensitive(params map[string]string, key string) (string, bool) {
	lowerKey := strings.ToLower(key)
	for k, v := range params {
		if strings.ToLower(k) == lowerKey {
			return v, true
		}
	}

	return "", false
}

// this function could be used also for other bindings
func setFieldValue(field reflect.Value, paramValue string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(paramValue)
	case reflect.Int:
		err := intField(paramValue, 0, field)
		printError(err, paramValue, "int")
	case reflect.Int8:
		err := intField(paramValue, 8, field)
		printError(err, paramValue, "int8")
	case reflect.Int16:
		err := intField(paramValue, 16, field)
		printError(err, paramValue, "int16")
	case reflect.Int32:
		err := intField(paramValue, 32, field)
		printError(err, paramValue, "int32")
	case reflect.Int64:
		err := intField(paramValue, 64, field)
		printError(err, paramValue, "uint64")
	case reflect.Uint:
		err := uintField(paramValue, 0, field)
		printError(err, paramValue, "uint")
	case reflect.Uint8:
		err := uintField(paramValue, 8, field)
		printError(err, paramValue, "uint8")
	case reflect.Uint16:
		err := uintField(paramValue, 16, field)
		printError(err, paramValue, "uint16")
	case reflect.Uint32:
		err := uintField(paramValue, 32, field)
		printError(err, paramValue, "uint32")
	case reflect.Uint64:
		err := uintField(paramValue, 64, field)
		printError(err, paramValue, "uint64")
	case reflect.Float32:
		err := floatField(paramValue, 32, field)
		printError(err, paramValue, "float32")
	case reflect.Float64:
		err := floatField(paramValue, 64, field)
		printError(err, paramValue, "float64")
	case reflect.Bool:
		err := boolField(paramValue, field)
		printError(err, paramValue, "bool")
	case reflect.Ptr:
		if !field.Elem().IsValid() {
			field.Set(reflect.New(field.Type().Elem()))
		}
		return setFieldValue(field.Elem(), paramValue)
	case reflect.Struct:
		switch field.Interface().(type) {
		case time.Time:
			err := timeField(paramValue, field)
			printError(err, paramValue, "time")
			return err
		}
		err := json.Unmarshal(StringToBytes(paramValue), field.Addr().Interface())
		printError(err, paramValue, "struct")
	case reflect.Map:
		err := json.Unmarshal(StringToBytes(paramValue), field.Addr().Interface())
		printError(err, paramValue, "map")
	default:
		if field.Type() == reflect.TypeOf((*multipart.FileHeader)(nil)) || field.Type() == reflect.TypeOf(multipart.FileHeader{}) {
			return fmt.Errorf("file header type not supported")
		}
	}

	return nil
}

func printError(err error, paramValue string, t string) {
	if err != nil {
		fmt.Printf("[ERROR] Error binding parameter %s, type %s \n", paramValue, t)
	}
}
