package binding

import (
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestValidator_Validate(t *testing.T) {
	v := Validator{}

	t.Run("NonPointer", func(t *testing.T) {
		type TestStruct struct {
			Name string `v:"required"`
		}
		s := TestStruct{Name: "test"}
		err := v.Validate(s)
		if err == nil {
			t.Error("Expected error for non-pointer input")
		}
		if !strings.Contains(err.Error(), "must be a non-nil pointer") {
			t.Errorf("Expected pointer error, got: %s", err.Error())
		}
	})

	t.Run("NilPointer", func(t *testing.T) {
		var s *struct{}
		err := v.Validate(s)
		if err == nil {
			t.Error("Expected error for nil pointer")
		}
		if !strings.Contains(err.Error(), "must be a non-nil pointer") {
			t.Errorf("Expected nil pointer error, got: %s", err.Error())
		}
	})

	t.Run("NonStruct", func(t *testing.T) {
		s := "not a struct"
		err := v.Validate(&s)
		if err == nil {
			t.Error("Expected error for non-struct input")
		}
		if !strings.Contains(err.Error(), "must be a pointer to a struct") {
			t.Errorf("Expected struct error, got: %s", err.Error())
		}
	})

	t.Run("RequiredValid", func(t *testing.T) {
		type TestStruct struct {
			Name string `v:"required"`
		}
		s := &TestStruct{Name: "test"}
		err := v.Validate(s)
		if err != nil {
			t.Errorf("Expected no error for valid required field, got: %s", err.Error())
		}
	})

	t.Run("RequiredInvalid", func(t *testing.T) {
		type TestStruct struct {
			Name string `v:"required"`
		}
		s := &TestStruct{}
		err := v.Validate(s)
		if err == nil {
			t.Error("Expected error for empty required field")
		}
		if !strings.Contains(err.Error(), "is required") {
			t.Errorf("Expected required error, got: %s", err.Error())
		}
	})

	t.Run("EmailValid", func(t *testing.T) {
		type TestStruct struct {
			Email string `v:"email"`
		}
		s := &TestStruct{Email: "test@example.com"}
		err := v.Validate(s)
		if err != nil {
			t.Errorf("Expected no error for valid email, got: %s", err.Error())
		}
	})

	t.Run("EmailInvalid", func(t *testing.T) {
		type TestStruct struct {
			Email string `v:"email"`
		}
		s := &TestStruct{Email: "invalid-email"}
		err := v.Validate(s)
		if err == nil {
			t.Error("Expected error for invalid email")
		}
		if !strings.Contains(err.Error(), "invalid email format") {
			t.Errorf("Expected email error, got: %s", err.Error())
		}
	})

	t.Run("UUIDValid", func(t *testing.T) {
		type TestStruct struct {
			ID string `v:"uuid"`
		}
		validUUID := uuid.New().String()
		s := &TestStruct{ID: validUUID}
		err := v.Validate(s)
		if err != nil {
			t.Errorf("Expected no error for valid UUID, got: %s", err.Error())
		}
	})

	t.Run("UUIDInvalid", func(t *testing.T) {
		type TestStruct struct {
			ID string `v:"uuid"`
		}
		s := &TestStruct{ID: "invalid-uuid"}
		err := v.Validate(s)
		if err == nil {
			t.Error("Expected error for invalid UUID")
		}
		if !strings.Contains(err.Error(), "invalid UUID format") {
			t.Errorf("Expected UUID error, got: %s", err.Error())
		}
	})

	t.Run("AlphanumericValid", func(t *testing.T) {
		type TestStruct struct {
			Code string `v:"alphanumeric"`
		}
		s := &TestStruct{Code: "abc123"}
		err := v.Validate(s)
		if err != nil {
			t.Errorf("Expected no error for valid alphanumeric, got: %s", err.Error())
		}
	})

	t.Run("AlphanumericInvalid", func(t *testing.T) {
		type TestStruct struct {
			Code string `v:"alphanumeric"`
		}
		s := &TestStruct{Code: "abc-123"}
		err := v.Validate(s)
		if err == nil {
			t.Error("Expected error for invalid alphanumeric")
		}
		if !strings.Contains(err.Error(), "must be alphanumeric") {
			t.Errorf("Expected alphanumeric error, got: %s", err.Error())
		}
	})

	t.Run("AlphaValid", func(t *testing.T) {
		type TestStruct struct {
			Name string `v:"alpha"`
		}
		s := &TestStruct{Name: "abcDEF"}
		err := v.Validate(s)
		if err != nil {
			t.Errorf("Expected no error for valid alpha, got: %s", err.Error())
		}
	})

	t.Run("AlphaInvalid", func(t *testing.T) {
		type TestStruct struct {
			Name string `v:"alpha"`
		}
		s := &TestStruct{Name: "abc123"}
		err := v.Validate(s)
		if err == nil {
			t.Error("Expected error for invalid alpha")
		}
		if !strings.Contains(err.Error(), "must be alphabetic") {
			t.Errorf("Expected alpha error, got: %s", err.Error())
		}
	})

	t.Run("NumericValid", func(t *testing.T) {
		type TestStruct struct {
			Code string `v:"numeric"`
		}
		s := &TestStruct{Code: "123456"}
		err := v.Validate(s)
		if err != nil {
			t.Errorf("Expected no error for valid numeric, got: %s", err.Error())
		}
	})

	t.Run("NumericInvalid", func(t *testing.T) {
		type TestStruct struct {
			Code string `v:"numeric"`
		}
		s := &TestStruct{Code: "123abc"}
		err := v.Validate(s)
		if err == nil {
			t.Error("Expected error for invalid numeric")
		}
		if !strings.Contains(err.Error(), "must be numeric") {
			t.Errorf("Expected numeric error, got: %s", err.Error())
		}
	})

	t.Run("DateTimeValid", func(t *testing.T) {
		type TestStruct struct {
			Date string `v:"datetime"`
		}
		s := &TestStruct{Date: "2023-01-01T12:00:00Z"}
		err := v.Validate(s)
		if err != nil {
			t.Errorf("Expected no error for valid datetime, got: %s", err.Error())
		}
	})

	t.Run("DateTimeInvalid", func(t *testing.T) {
		type TestStruct struct {
			Date string `v:"datetime"`
		}
		s := &TestStruct{Date: "invalid-date"}
		err := v.Validate(s)
		if err == nil {
			t.Error("Expected error for invalid datetime")
		}
		if !strings.Contains(err.Error(), "invalid datetime format") {
			t.Errorf("Expected datetime error, got: %s", err.Error())
		}
	})

	t.Run("LengthValid", func(t *testing.T) {
		type TestStruct struct {
			Code string `v:"len=5"`
		}
		s := &TestStruct{Code: "12345"}
		err := v.Validate(s)
		if err != nil {
			t.Errorf("Expected no error for valid length, got: %s", err.Error())
		}
	})

	t.Run("LengthInvalid", func(t *testing.T) {
		type TestStruct struct {
			Code string `v:"len=5"`
		}
		s := &TestStruct{Code: "123"}
		err := v.Validate(s)
		if err == nil {
			t.Error("Expected error for invalid length")
		}
		if !strings.Contains(err.Error(), "length must be 5") {
			t.Errorf("Expected length error, got: %s", err.Error())
		}
	})

	t.Run("MinLengthValid", func(t *testing.T) {
		type TestStruct struct {
			Password string `v:"minlen=8"`
		}
		s := &TestStruct{Password: "password123"}
		err := v.Validate(s)
		if err != nil {
			t.Errorf("Expected no error for valid min length, got: %s", err.Error())
		}
	})

	t.Run("MinLengthInvalid", func(t *testing.T) {
		type TestStruct struct {
			Password string `v:"minlen=8"`
		}
		s := &TestStruct{Password: "pass"}
		err := v.Validate(s)
		if err == nil {
			t.Error("Expected error for invalid min length")
		}
		if !strings.Contains(err.Error(), "length must be at least 8") {
			t.Errorf("Expected min length error, got: %s", err.Error())
		}
	})

	t.Run("MaxLengthValid", func(t *testing.T) {
		type TestStruct struct {
			Username string `v:"maxlen=10"`
		}
		s := &TestStruct{Username: "user123"}
		err := v.Validate(s)
		if err != nil {
			t.Errorf("Expected no error for valid max length, got: %s", err.Error())
		}
	})

	t.Run("MaxLengthInvalid", func(t *testing.T) {
		type TestStruct struct {
			Username string `v:"maxlen=5"`
		}
		s := &TestStruct{Username: "verylongusername"}
		err := v.Validate(s)
		if err == nil {
			t.Error("Expected error for invalid max length")
		}
		if !strings.Contains(err.Error(), "length must be at most 5") {
			t.Errorf("Expected max length error, got: %s", err.Error())
		}
	})

	t.Run("NumericComparisonInt", func(t *testing.T) {
		type TestStruct struct {
			Age    int     `v:"gte=18;lte=65"`
			Score  float64 `v:"gt=0;lt=100"`
			Rating uint    `v:"gte=1;lte=5"`
		}
		s := &TestStruct{Age: 25, Score: 85.5, Rating: 4}
		err := v.Validate(s)
		if err != nil {
			t.Errorf("Expected no error for valid numeric comparisons, got: %s", err.Error())
		}
	})

	t.Run("EqualityString", func(t *testing.T) {
		type TestStruct struct {
			Status string `v:"eq=active"`
		}
		s := &TestStruct{Status: "active"}
		err := v.Validate(s)
		if err != nil {
			t.Errorf("Expected no error for valid equality, got: %s", err.Error())
		}
	})

	t.Run("EqualityInvalid", func(t *testing.T) {
		type TestStruct struct {
			Status string `v:"eq=active"`
		}
		s := &TestStruct{Status: "inactive"}
		err := v.Validate(s)
		if err == nil {
			t.Error("Expected error for invalid equality")
		}
		if !strings.Contains(err.Error(), "must be equal to active") {
			t.Errorf("Expected equality error, got: %s", err.Error())
		}
	})

	t.Run("NotEqualValid", func(t *testing.T) {
		type TestStruct struct {
			Status string `v:"neq=banned"`
		}
		s := &TestStruct{Status: "active"}
		err := v.Validate(s)
		if err != nil {
			t.Errorf("Expected no error for valid not equal, got: %s", err.Error())
		}
	})

	t.Run("RegexValid", func(t *testing.T) {
		type TestStruct struct {
			Phone string `v:"regex=^\\+[1-9]\\d{1,14}$"`
		}
		s := &TestStruct{Phone: "+1234567890"}
		err := v.Validate(s)
		if err != nil {
			t.Errorf("Expected no error for valid regex, got: %s", err.Error())
		}
	})

	t.Run("RegexInvalid", func(t *testing.T) {
		type TestStruct struct {
			Phone string `v:"regex=^\\+[1-9]\\d{1,14}$"`
		}
		s := &TestStruct{Phone: "invalid-phone"}
		err := v.Validate(s)
		if err == nil {
			t.Error("Expected error for invalid regex")
		}
		if !strings.Contains(err.Error(), "does not match the regex pattern") {
			t.Errorf("Expected regex error, got: %s", err.Error())
		}
	})

	t.Run("OneOfValid", func(t *testing.T) {
		type TestStruct struct {
			Color string `v:"oneof=red blue green"`
		}
		s := &TestStruct{Color: "blue"}
		err := v.Validate(s)
		if err != nil {
			t.Errorf("Expected no error for valid oneof, got: %s", err.Error())
		}
	})

	t.Run("OneOfInvalid", func(t *testing.T) {
		type TestStruct struct {
			Color string `v:"oneof=red blue green"`
		}
		s := &TestStruct{Color: "yellow"}
		err := v.Validate(s)
		if err == nil {
			t.Error("Expected error for invalid oneof")
		}
		if !strings.Contains(err.Error(), "must be one of: red blue green") {
			t.Errorf("Expected oneof error, got: %s", err.Error())
		}
	})

	t.Run("NotOneOfValid", func(t *testing.T) {
		type TestStruct struct {
			Username string `v:"notoneof=admin root system"`
		}
		s := &TestStruct{Username: "johndoe"}
		err := v.Validate(s)
		if err != nil {
			t.Errorf("Expected no error for valid username, got: %s", err.Error())
		}
	})

	t.Run("ContainsValid", func(t *testing.T) {
		type TestStruct struct {
			Description string `v:"contains=important"`
		}
		s := &TestStruct{Description: "This is an important message"}
		err := v.Validate(s)
		if err != nil {
			t.Errorf("Expected no error for valid contains, got: %s", err.Error())
		}
	})

	t.Run("StartsWithValid", func(t *testing.T) {
		type TestStruct struct {
			URL string `v:"startswith=https://"`
		}
		s := &TestStruct{URL: "https://example.com"}
		err := v.Validate(s)
		if err != nil {
			t.Errorf("Expected no error for valid startswith, got: %s", err.Error())
		}
	})

	t.Run("EndsWithValid", func(t *testing.T) {
		type TestStruct struct {
			Filename string `v:"endswith=.pdf"`
		}
		s := &TestStruct{Filename: "document.pdf"}
		err := v.Validate(s)
		if err != nil {
			t.Errorf("Expected no error for valid endswith, got: %s", err.Error())
		}
	})

	t.Run("MultipleRulesValid", func(t *testing.T) {
		type TestStruct struct {
			Email    string `v:"required;email"`
			Password string `v:"required;minlen=8;contains=@"`
			Age      int    `v:"gte=18;lte=120"`
		}
		s := &TestStruct{
			Email:    "user@example.com",
			Password: "password@123",
			Age:      25,
		}
		err := v.Validate(s)
		if err != nil {
			t.Errorf("Expected no error for valid multiple rules, got: %s", err.Error())
		}
	})

	t.Run("InvalidRuleFormat", func(t *testing.T) {
		type TestStruct struct {
			Field string `v:"invalid=rule=format"`
		}
		s := &TestStruct{Field: "test"}
		err := v.Validate(s)
		if err == nil {
			t.Error("Expected error for invalid rule format")
		}
		if !strings.Contains(err.Error(), "invalid validation rule format") {
			t.Errorf("Expected rule format error, got: %s", err.Error())
		}
	})

	t.Run("UnknownRule", func(t *testing.T) {
		type TestStruct struct {
			Field string `v:"unknownrule"`
		}
		s := &TestStruct{Field: "test"}
		err := v.Validate(s)
		if err != nil {
			t.Errorf("Expected no error for unknown rule (should be warned), got: %s", err.Error())
		}
	})

	t.Run("EmptyValidationTag", func(t *testing.T) {
		type TestStruct struct {
			Field string `v:""`
		}
		s := &TestStruct{Field: "test"}
		err := v.Validate(s)
		if err != nil {
			t.Errorf("Expected no error for empty validation tag, got: %s", err.Error())
		}
	})

	t.Run("NoValidationTag", func(t *testing.T) {
		type TestStruct struct {
			Field string
		}
		s := &TestStruct{Field: "test"}
		err := v.Validate(s)
		if err != nil {
			t.Errorf("Expected no error for no validation tag, got: %s", err.Error())
		}
	})
}

func TestCheckDateTime(t *testing.T) {
	t.Run("ValidDateTime", func(t *testing.T) {
		field := reflect.ValueOf("2023-01-01T12:00:00Z")
		err := checkDateTime(field, "2006-01-02T15:04:05Z07:00")
		if err != nil {
			t.Errorf("Expected no error for valid datetime, got: %s", err.Error())
		}
	})

	t.Run("InvalidDateTime", func(t *testing.T) {
		field := reflect.ValueOf("invalid-date")
		err := checkDateTime(field, "2006-01-02T15:04:05Z07:00")
		if err == nil {
			t.Error("Expected error for invalid datetime")
		}
	})

	t.Run("EmptyDateTime", func(t *testing.T) {
		field := reflect.ValueOf("")
		err := checkDateTime(field, "2006-01-02T15:04:05Z07:00")
		if err == nil {
			t.Error("Expected error for empty datetime")
		}
	})

	t.Run("NonStringField", func(t *testing.T) {
		field := reflect.ValueOf(123)
		err := checkDateTime(field, "2006-01-02T15:04:05Z07:00")
		if err == nil {
			t.Error("Expected error for non-string field")
		}
	})
}

func TestHelperFunctions(t *testing.T) {
	t.Run("IsNumeric", func(t *testing.T) {
		if !isNumeric('5') {
			t.Error("Expected '5' to be numeric")
		}
		if isNumeric('a') {
			t.Error("Expected 'a' to not be numeric")
		}
	})

	t.Run("IsAlpha", func(t *testing.T) {
		if !isAlpha('A') {
			t.Error("Expected 'A' to be alpha")
		}
		if !isAlpha('z') {
			t.Error("Expected 'z' to be alpha")
		}
		if isAlpha('5') {
			t.Error("Expected '5' to not be alpha")
		}
	})

	t.Run("IsAlphanumeric", func(t *testing.T) {
		if !isAlphanumeric('A') {
			t.Error("Expected 'A' to be alphanumeric")
		}
		if !isAlphanumeric('5') {
			t.Error("Expected '5' to be alphanumeric")
		}
		if isAlphanumeric('@') {
			t.Error("Expected '@' to not be alphanumeric")
		}
	})
}

func TestCheckEqualType(t *testing.T) {
	t.Run("StringEqual", func(t *testing.T) {
		field := reflect.ValueOf("test")
		if !checkEqualType(field, "test") {
			t.Error("Expected string equality to pass")
		}
		if checkEqualType(field, "different") {
			t.Error("Expected string inequality to fail")
		}
	})

	t.Run("IntEqual", func(t *testing.T) {
		field := reflect.ValueOf(42)
		if !checkEqualType(field, "42") {
			t.Error("Expected int equality to pass")
		}
		if checkEqualType(field, "24") {
			t.Error("Expected int inequality to fail")
		}
	})

	t.Run("BoolEqual", func(t *testing.T) {
		field := reflect.ValueOf(true)
		if !checkEqualType(field, "true") {
			t.Error("Expected bool equality to pass")
		}
		if checkEqualType(field, "false") {
			t.Error("Expected bool inequality to fail")
		}
	})

	t.Run("Float64Equal", func(t *testing.T) {
		field := reflect.ValueOf(3.14)
		if !checkEqualType(field, "3.14") {
			t.Error("Expected float equality to pass")
		}
		if checkEqualType(field, "2.71") {
			t.Error("Expected float inequality to fail")
		}
	})
}

func TestCheckNumericComparison(t *testing.T) {
	t.Run("IntComparison", func(t *testing.T) {
		field := reflect.ValueOf(10)
		err := checkNumericComparison(field, "5", func(f, v float64) bool { return f > v })
		if err != nil {
			t.Errorf("Expected no error for valid int comparison, got: %s", err.Error())
		}
	})

	t.Run("UnsupportedType", func(t *testing.T) {
		field := reflect.ValueOf("string")
		err := checkNumericComparison(field, "5", func(f, v float64) bool { return f > v })
		if err == nil {
			t.Error("Expected error for unsupported type")
		}
	})

	t.Run("InvalidRuleValue", func(t *testing.T) {
		field := reflect.ValueOf(10)
		err := checkNumericComparison(field, "invalid", func(f, v float64) bool { return f > v })
		if err == nil {
			t.Error("Expected error for invalid rule value")
		}
	})
}
