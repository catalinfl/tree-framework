package binding

import (
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"reflect"
	"strings"
)

type QueryBinding struct{}

func (QueryBinding) Name() string {
	return "query"
}

func (QueryBinding) BindQuery(req *http.Request, v any) error {
	if req == nil || req.URL.Query() == nil {
		return errors.New("request is invalid")
	}

	return decodeQuery(req.URL.Query(), v)
}

func decodeQuery(query map[string][]string, v any) error {
	fmt.Println(query)
	err := bindMap(query, v)
	if err != nil {
		return errors.New("error decoding query: " + err.Error())
	}
	return validator.Validate(v)
}

func bindMap(m map[string][]string, v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("v must be a pointer")
	}

	elem := rv.Elem()
	if elem.Kind() != reflect.Struct {
		return errors.New("v must be a struct pointer")
	}

	structType := elem.Type()

	for i := 0; i < structType.NumField(); i++ {
		fieldVal := elem.Field(i)
		fieldType := structType.Field(i)

		if !fieldVal.CanSet() {
			continue
		}

		queryTag := fieldType.Tag.Get("query")
		if queryTag == "" {
			fmt.Printf("[WARNING] Query tag not found for field: %s. Used field name for this.", fieldType.Name)
			continue
		}

		paramValues, found := findMultipleParamCaseInsensitive(m, queryTag)
		if !found || len(paramValues) == 0 {
			continue
		}

		if err := setQueryField(fieldVal, paramValues); err != nil {
			return fmt.Errorf("error setting field %s: %w", fieldType.Name, err)
		}
	}

	return nil
}

func setQueryField(fieldVal reflect.Value, paramValues []string) error {
	switch fieldVal.Kind() {
	case reflect.String:
		fieldVal.SetString(paramValues[0])
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intField(paramValues[0], fieldVal.Type().Bits(), fieldVal)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintField(paramValues[0], fieldVal.Type().Bits(), fieldVal)
	case reflect.Float32, reflect.Float64:
		floatField(paramValues[0], fieldVal.Type().Bits(), fieldVal)
	case reflect.Bool:
		boolField(paramValues[0], fieldVal)
	case reflect.Slice:
		elemType := fieldVal.Type().Elem()
		slice := reflect.MakeSlice(fieldVal.Type(), 0, len(paramValues))
		for _, paramValue := range paramValues {
			elemValPtr := reflect.New(elemType)
			if err := setFieldValue(elemValPtr.Elem(), paramValue); err != nil {
				return fmt.Errorf("error setting field %s: %w", fieldVal.Type().Name(), err)
			}
			slice = reflect.Append(slice, elemValPtr.Elem())
		}
		fieldVal.Set(slice)
	case reflect.Array:
		arrLen := fieldVal.Len()
		if len(paramValues) != arrLen {
			return fmt.Errorf("array length mismatch: expected %d, got %d", arrLen, len(paramValues))
		}
		elemType := fieldVal.Type().Elem()
		switch elemType.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			for i := 0; i < arrLen; i++ {
				arrElem := fieldVal.Index(i)
				if err := intField(paramValues[i], arrElem.Type().Bits(), arrElem); err != nil {
					return fmt.Errorf("error setting array field %s: %w", fieldVal.Type().Name(), err)
				}
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			for i := 0; i < arrLen; i++ {
				arrElem := fieldVal.Index(i)
				if err := uintField(paramValues[i], arrElem.Type().Bits(), arrElem); err != nil {
					return fmt.Errorf("error setting array field %s: %w", fieldVal.Type().Name(), err)
				}
			}
		case reflect.Float32, reflect.Float64:
			for i := 0; i < arrLen; i++ {
				arrElem := fieldVal.Index(i)
				if err := floatField(paramValues[i], arrElem.Type().Bits(), arrElem); err != nil {
					return fmt.Errorf("error setting array field %s: %w", fieldVal.Type().Name(), err)
				}
			}
		case reflect.String:
			for i := 0; i < arrLen; i++ {
				arrElem := fieldVal.Index(i)
				arrElem.SetString(paramValues[i])
			}
		default:
			if fieldVal.Type() == reflect.TypeOf((*multipart.FileHeader)(nil)) {
				return nil
			}
			if len(paramValues) > 0 {
				if err := setFieldValue(fieldVal, paramValues[0]); err != nil {
					return fmt.Errorf("error setting field %s: %w", fieldVal.Type().Name(), err)
				}
			}
			return nil
		}
	}
	return nil
}

func findMultipleParamCaseInsensitive(m map[string][]string, key string) ([]string, bool) {
	for k, v := range m {
		if strings.EqualFold(k, key) {
			return v, true
		}
	}
	return nil, false
}
