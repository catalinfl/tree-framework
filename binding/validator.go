package binding

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// validation should be done here
type Validator struct{}

var validator Validator = Validator{}

func (Validator) Validate(v any) error {
	rv := reflect.ValueOf(v)

	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("binding destination must be a non-nil pointer")
	}

	structVal := rv.Elem()
	if structVal.Kind() != reflect.Struct {
		return fmt.Errorf("binding destination must be a pointer to a struct")
	}

	structType := structVal.Type()

	for i := 0; i < structType.NumField(); i++ {
		fieldVal := structVal.Field(i)
		fieldType := structType.Field(i)

		validationTag := fieldType.Tag.Get("v")
		if validationTag != "" {
			if err := validateField(fieldVal, validationTag); err != nil {
				return fmt.Errorf("validation error for field %s: %w", fieldType.Name, err)
			}
			fmt.Println("Validation passed for field:", fieldType.Name)
		}
	}

	return nil
}

func validateField(field reflect.Value, tag string) error {
	if tag == "" {
		return nil
	}

	rules := strings.Split(tag, ";")
	for _, rule := range rules {
		parts := strings.Split(rule, "=")
		if len(parts) == 1 {
			ruleName := strings.TrimSpace(rule)
			switch ruleName {
			case "uuid":
				if err := checkUUID(field); err != nil {
					return fmt.Errorf("invalid UUID format (%w)", err)
				}
			case "email":
				if err := checkEmail(field); err != nil {
					return fmt.Errorf("invalid email format (%w)", err)
				}
			case "alphanumeric":
				if !alphanumeric(field) {
					return fmt.Errorf("must be alphanumeric")
				}
			case "alpha":
				if !alpha(field) {
					return fmt.Errorf("must be alphabetic")
				}
			case "numeric":
				if !numeric(field) {
					return fmt.Errorf("must be numeric")
				}
			case "datetime":
				if err := checkDateTime(field, "2006-01-02T15:04:05Z07:00"); err != nil {
					return fmt.Errorf("invalid datetime format (%w)", err)
				}
			case "required":
				if field.IsZero() {
					return fmt.Errorf("is required")
				}
			}

			continue
		} else if len(parts) != 2 {
			return fmt.Errorf("invalid validation rule format: %s", rule)
		}

		ruleName := strings.TrimSpace(parts[0])
		ruleValueStr := strings.TrimSpace(parts[1])

		switch ruleName {
		case "lte":
			if err := checkNumericComparison(field, ruleValueStr, func(f, v float64) bool { return f <= v }); err != nil {
				return fmt.Errorf("must be less than or equal to %s (%w)", ruleValueStr, err)
			}
		case "gte":
			if err := checkNumericComparison(field, ruleValueStr, func(f, v float64) bool { return f >= v }); err != nil {
				return fmt.Errorf("must be greater than or equal to %s (%w)", ruleValueStr, err)
			}
		case "gt":
			if err := checkNumericComparison(field, ruleValueStr, func(f, v float64) bool { return f > v }); err != nil {
				return fmt.Errorf("must be greater than %s (%w)", ruleValueStr, err)
			}
		case "lt":
			if err := checkNumericComparison(field, ruleValueStr, func(f, v float64) bool { return f < v }); err != nil {
				return fmt.Errorf("must be less than %s (%w)", ruleValueStr, err)
			}
		case "len":
			if err := checkEqualLength(field, ruleValueStr); err != nil {
				return fmt.Errorf("length must be %s (%w)", ruleValueStr, err)
			}
		case "minlen":
			err := checkLengthComparison(field, ruleValueStr, func(f, v int) bool { return f >= v })
			if err != nil {
				return fmt.Errorf("length must be at least %s (%w)", ruleValueStr, err)
			}
		case "maxlen":
			err := checkLengthComparison(field, ruleValueStr, func(f, v int) bool { return f <= v })
			if err != nil {
				return fmt.Errorf("length must be at most %s (%w)", ruleValueStr, err)
			}
		case "eq":
			// usage for string, int, uint, float, bool
			if b := checkEqualType(field, ruleValueStr); !b {
				return fmt.Errorf("must be equal to %s", ruleValueStr)
			}
		case "neq":
			if b := checkEqualType(field, ruleValueStr); b {
				return fmt.Errorf("must not be equal to %s", ruleValueStr)
			}
		case "regex":
			if err := checkRegex(field, ruleValueStr); err != nil {
				return fmt.Errorf("value does not match the regex pattern (%w)", err)
			}
		case "oneof":
			if err := checkOneOf(field, ruleValueStr); err != nil {
				return fmt.Errorf("value must be one of: %s (%w)", ruleValueStr, err)
			}
		case "notoneof":
			if err := checkNotOneOf(field, ruleValueStr); err != nil {
				return fmt.Errorf("value must not be one of: %s (%w)", ruleValueStr, err)
			}
		case "contains":
			if err := checkContains(field, ruleValueStr); err != nil {
				return fmt.Errorf("value must contain: %s (%w)", ruleValueStr, err)
			}
		case "startswith":
			if err := checkStartsWith(field, ruleValueStr); err != nil {
				return fmt.Errorf("value must start with: %s (%w)", ruleValueStr, err)
			}
		case "endswith":
			if err := checkEndsWith(field, ruleValueStr); err != nil {
				return fmt.Errorf("value must end with: %s (%w)", ruleValueStr, err)
			}
		default:
			fmt.Printf("[Warning] Unknown validation rule: %s\n", ruleName)
		}
	}
	return nil
}

func checkDateTime(field reflect.Value, ruleValueStr string) error {
	if field.Kind() != reflect.String {
		return fmt.Errorf("unsupported field type for datetime check: %s", field.Kind())
	}

	datetimeStr := field.String()
	if datetimeStr == "" {
		return fmt.Errorf("datetime string is empty")
	}

	_, err := time.Parse(ruleValueStr, datetimeStr)
	if err != nil {
		return fmt.Errorf("invalid datetime format: %w", err)
	}

	return nil
}

func alphanumeric(field reflect.Value) bool {
	if field.Kind() != reflect.String {
		return false
	}
	value := field.String()
	for _, char := range value {
		if !isAlphanumeric(char) {
			return false
		}
	}
	return true
}

func alpha(field reflect.Value) bool {
	if field.Kind() != reflect.String {
		return false
	}
	value := field.String()
	for _, char := range value {
		if !isAlpha(char) {
			return false
		}
	}
	return true
}

func numeric(field reflect.Value) bool {
	if field.Kind() != reflect.String {
		return false
	}
	value := field.String()
	for _, char := range value {
		if !isNumeric(char) {
			return false
		}
	}
	return true
}

func isNumeric(r rune) bool {
	if r >= '0' && r <= '9' {
		return true
	}
	return false
}

func isAlpha(r rune) bool {
	if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
		return true
	}
	return false
}

func isAlphanumeric(r rune) bool {
	if (r >= '0' && r <= '9') || (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
		return true
	}
	return false
}

func oneOf(field reflect.Value, ruleValueStr string) bool {
	options := strings.Split(ruleValueStr, " ")
	value := field.String()
	for _, option := range options {
		if strings.TrimSpace(option) == value {
			return true
		}
	}

	return false
}

func checkUUID(field reflect.Value) error {
	if field.Kind() != reflect.String {
		return fmt.Errorf("unsupported field type for UUID check: %s", field.Kind())
	}

	uuidStr := field.String()

	_, err := uuid.Parse(uuidStr)
	if err != nil {
		return fmt.Errorf("invalid UUID format: %w", err)
	}

	return nil
}

func checkEndsWith(field reflect.Value, ruleValueStr string) error {
	if field.Kind() != reflect.String {
		return fmt.Errorf("unsupported field type for endswith check: %s", field.Kind())
	}
	if strings.HasSuffix(field.String(), ruleValueStr) {
		return nil
	}
	return fmt.Errorf("value must end with: %s", ruleValueStr)
}

func checkStartsWith(field reflect.Value, ruleValueStr string) error {
	if field.Kind() != reflect.String {
		return fmt.Errorf("unsupported field type for startswith check: %s", field.Kind())
	}
	if strings.HasPrefix(field.String(), ruleValueStr) {
		return nil
	}
	return fmt.Errorf("value must start with: %s", ruleValueStr)
}

func checkContains(field reflect.Value, ruleValueStr string) error {
	if field.Kind() != reflect.String {
		return fmt.Errorf("unsupported field type for contains check: %s", field.Kind())
	}
	if strings.Contains(field.String(), ruleValueStr) {
		return nil
	}
	return fmt.Errorf("value must contain: %s", ruleValueStr)
}

func checkNotOneOf(field reflect.Value, ruleValueStr string) error {
	if field.Kind() != reflect.String {
		return fmt.Errorf("unsupported field type for notoneof check: %s", field.Kind())
	}
	if !oneOf(field, ruleValueStr) {
		return nil
	}

	return fmt.Errorf("value must NOT be one of: %s", ruleValueStr) // FIXED MESSAGE
}

func checkOneOf(field reflect.Value, ruleValueStr string) error {
	if field.Kind() != reflect.String {
		return fmt.Errorf("unsupported field type for oneof check: %s", field.Kind())
	}
	if oneOf(field, ruleValueStr) {
		return nil
	}
	return fmt.Errorf("value must be one of: %s", ruleValueStr)
}

func checkRegex(field reflect.Value, ruleValueStr string) error {
	if field.Kind() != reflect.String {
		return fmt.Errorf("unsupported field type for regex check: %s", field.Kind())
	}

	regex, err := regexp.Compile(ruleValueStr)
	if err != nil {
		return fmt.Errorf("invalid regex pattern: %w", err)
	}

	if !regex.MatchString(field.String()) {
		return fmt.Errorf("value does not match the regex pattern")
	}

	return nil
}

func checkEmail(field reflect.Value) error {
	if field.Kind() != reflect.String {
		return fmt.Errorf("unsupported field type for email check: %s", field.Kind())
	}

	email := field.String()
	if len(email) < 3 || len(email) > 254 {
		return fmt.Errorf("email length must be between 3 and 254 characters")
	}

	t, err := regexp.MatchString(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`, email)
	if err != nil {
		return fmt.Errorf("error matching email regex: %w", err)
	}

	if !t {
		return fmt.Errorf("invalid email format")
	}

	return nil
}

func checkEqualType(field reflect.Value, ruleValueStr string) bool {
	if field.Kind() == reflect.String {
		return field.String() == ruleValueStr
	}

	if field.Kind() == reflect.Int || field.Kind() == reflect.Int8 || field.Kind() == reflect.Int16 || field.Kind() == reflect.Int32 || field.Kind() == reflect.Int64 {
		fieldInt, err := strconv.Atoi(ruleValueStr)
		if err != nil {
			return false
		}
		return field.Int() == int64(fieldInt)
	}

	if field.Kind() == reflect.Uint || field.Kind() == reflect.Uint8 || field.Kind() == reflect.Uint16 || field.Kind() == reflect.Uint32 || field.Kind() == reflect.Uint64 {
		fieldUint, err := strconv.ParseUint(ruleValueStr, 10, 64)
		if err != nil {
			return false
		}
		return field.Uint() == fieldUint
	}

	if field.Kind() == reflect.Float32 || field.Kind() == reflect.Float64 {
		fieldFloat, err := strconv.ParseFloat(ruleValueStr, 64)
		if err != nil {
			return false
		}
		return field.Float() == fieldFloat
	}

	if field.Kind() == reflect.Bool {
		fieldBool, err := strconv.ParseBool(ruleValueStr)
		if err != nil {
			return false
		}
		return field.Bool() == fieldBool
	}

	return false
}

func checkEqualLength(field reflect.Value, ruleValueStr string) error {
	if field.Kind() != reflect.String && field.Kind() != reflect.Slice && field.Kind() != reflect.Array {
		return fmt.Errorf("unsupported field type for length check: %s", field.Kind())
	}

	expectedLen, err := strconv.Atoi(ruleValueStr)
	if err != nil {
		return fmt.Errorf("invalid length value: %s", ruleValueStr)
	}

	fieldLen := field.Len()
	if fieldLen != expectedLen {
		return fmt.Errorf("length must be %d, got %d", expectedLen, fieldLen)
	}

	return nil
}

func checkLengthComparison(field reflect.Value, ruleValueStr string, compare func(f, v int) bool) error {
	var fieldLen int
	expectedLen, err := strconv.Atoi(ruleValueStr)
	if err != nil {
		return fmt.Errorf("invalid length value: %s", ruleValueStr)
	}

	if field.Kind() == reflect.String || field.Kind() == reflect.Slice || field.Kind() == reflect.Array {
		fieldLen = field.Len()
	} else {
		return fmt.Errorf("unsupported field type for length check: %s", field.Kind())
	}

	if !compare(fieldLen, expectedLen) {
		return fmt.Errorf("length %d does not satisfy the comparison with %d", fieldLen, expectedLen)
	}

	return nil
}

func checkNumericComparison(field reflect.Value, ruleValueStr string, compare func(f, v float64) bool) error {
	var fieldFloat float64
	var err error

	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		fieldFloat = float64(field.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		fieldFloat = float64(field.Uint())
	case reflect.Float32, reflect.Float64:
		fieldFloat = field.Float()
	default:
		return fmt.Errorf("unsupported field type for numeric comparison: %s", field.Kind())
	}

	ruleValueFloat, err := strconv.ParseFloat(ruleValueStr, 64)
	if err != nil {
		return fmt.Errorf("invalid rule value for numeric comparison: %s", ruleValueStr)
	}

	if !compare(fieldFloat, ruleValueFloat) {
		return fmt.Errorf("value %f does not satisfy the comparison with %f", fieldFloat, ruleValueFloat)
	}

	return nil
}
