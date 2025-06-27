// without memory allocation

package binding

import (
	"fmt"
	"reflect"
	"strconv"
	"time"
	"unsafe"
)

func StringToBytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

func BytesToString(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}

func intField(val string, bitSize int, fieldVal reflect.Value) error {
	if val == "" {
		val = "0"
	}

	i, err := strconv.ParseInt(val, 10, bitSize)
	if err != nil {
		return err
	}

	fieldVal.SetInt(i)
	return nil
}

func uintField(val string, bitSize int, fieldVal reflect.Value) error {
	if val == "" {
		val = "0"
	}

	i, err := strconv.ParseUint(val, 10, bitSize)
	if err != nil {
		return err
	}

	fieldVal.SetUint(i)
	return nil
}

func floatField(val string, bitSize int, fieldVal reflect.Value) error {
	if val == "" {
		val = "0.0"
	}

	f, err := strconv.ParseFloat(val, bitSize)
	if err != nil {
		return err
	}

	fieldVal.SetFloat(f)
	return nil
}

func boolField(val string, fieldVal reflect.Value) error {
	if val == "" {
		val = "false"
	}

	b, err := strconv.ParseBool(val)
	if err != nil {
		return err
	}

	fieldVal.SetBool(b)
	return nil
}

func timeField(val string, fieldVal reflect.Value) error {
	if fieldVal.Type() != reflect.TypeOf(time.Time{}) {
		return fmt.Errorf("expected time.Time, got %s", fieldVal.Type())
	}

	layout := time.RFC3339
	if val == "" {
		val = time.Now().Format(layout)
	}

	t, err := time.Parse(layout, val)
	if err != nil {
		return err
	}

	if !fieldVal.CanSet() {
		return fmt.Errorf("cannot set field value")
	}

	fieldVal.Set(reflect.ValueOf(t))
	return nil
}
