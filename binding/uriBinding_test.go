package binding

import (
	"reflect"
	"testing"
	"time"
)

func TestURIBinding_Name(t *testing.T) {
	binding := URIBinding{}
	if binding.Name() != "uri" {
		t.Errorf("Expected name 'uri', got '%s'", binding.Name())
	}
}

func TestURIBinding_BindURI(t *testing.T) {
	binding := URIBinding{}

	t.Run("BasicStringBinding", func(t *testing.T) {
		type User struct {
			ID   string `uri:"id"`
			Name string `uri:"name"`
		}

		params := map[string]string{
			"id":   "123",
			"name": "john",
		}

		var user User
		err := binding.BindURI(params, &user)

		if err != nil {
			t.Errorf("Expected no error, got: %s", err.Error())
		}

		if user.ID != "123" {
			t.Errorf("Expected ID '123', got '%s'", user.ID)
		}

		if user.Name != "john" {
			t.Errorf("Expected Name 'john', got '%s'", user.Name)
		}
	})

	t.Run("IntegerTypes", func(t *testing.T) {
		type Numbers struct {
			Int8Value  int8  `uri:"int8"`
			Int16Value int16 `uri:"int16"`
			Int32Value int32 `uri:"int32"`
			Int64Value int64 `uri:"int64"`
			IntValue   int   `uri:"int"`
		}

		params := map[string]string{
			"int8":  "127",
			"int16": "32767",
			"int32": "2147483647",
			"int64": "9223372036854775807",
			"int":   "42",
		}

		var numbers Numbers
		err := binding.BindURI(params, &numbers)

		if err != nil {
			t.Errorf("Expected no error, got: %s", err.Error())
		}

		if numbers.Int8Value != 127 {
			t.Errorf("Expected Int8Value 127, got %d", numbers.Int8Value)
		}

		if numbers.Int16Value != 32767 {
			t.Errorf("Expected Int16Value 32767, got %d", numbers.Int16Value)
		}

		if numbers.Int32Value != 2147483647 {
			t.Errorf("Expected Int32Value 2147483647, got %d", numbers.Int32Value)
		}

		if numbers.Int64Value != 9223372036854775807 {
			t.Errorf("Expected Int64Value 9223372036854775807, got %d", numbers.Int64Value)
		}

		if numbers.IntValue != 42 {
			t.Errorf("Expected IntValue 42, got %d", numbers.IntValue)
		}
	})

	t.Run("UnsignedIntegerTypes", func(t *testing.T) {
		type UnsignedNumbers struct {
			Uint8Value  uint8  `uri:"uint8"`
			Uint16Value uint16 `uri:"uint16"`
			Uint32Value uint32 `uri:"uint32"`
			Uint64Value uint64 `uri:"uint64"`
			UintValue   uint   `uri:"uint"`
		}

		params := map[string]string{
			"uint8":  "255",
			"uint16": "65535",
			"uint32": "4294967295",
			"uint64": "18446744073709551615",
			"uint":   "42",
		}

		var numbers UnsignedNumbers
		err := binding.BindURI(params, &numbers)

		if err != nil {
			t.Errorf("Expected no error, got: %s", err.Error())
		}

		if numbers.Uint8Value != 255 {
			t.Errorf("Expected Uint8Value 255, got %d", numbers.Uint8Value)
		}

		if numbers.Uint16Value != 65535 {
			t.Errorf("Expected Uint16Value 65535, got %d", numbers.Uint16Value)
		}

		if numbers.Uint32Value != 4294967295 {
			t.Errorf("Expected Uint32Value 4294967295, got %d", numbers.Uint32Value)
		}

		if numbers.Uint64Value != 18446744073709551615 {
			t.Errorf("Expected Uint64Value 18446744073709551615, got %d", numbers.Uint64Value)
		}

		if numbers.UintValue != 42 {
			t.Errorf("Expected UintValue 42, got %d", numbers.UintValue)
		}
	})

	t.Run("FloatTypes", func(t *testing.T) {
		type Floats struct {
			Float32Value float32 `uri:"float32"`
			Float64Value float64 `uri:"float64"`
		}

		params := map[string]string{
			"float32": "3.14",
			"float64": "2.718281828",
		}

		var floats Floats
		err := binding.BindURI(params, &floats)

		if err != nil {
			t.Errorf("Expected no error, got: %s", err.Error())
		}

		if floats.Float32Value != 3.14 {
			t.Errorf("Expected Float32Value 3.14, got %f", floats.Float32Value)
		}

		if floats.Float64Value != 2.718281828 {
			t.Errorf("Expected Float64Value 2.718281828, got %f", floats.Float64Value)
		}
	})

	t.Run("BooleanType", func(t *testing.T) {
		type Flags struct {
			IsActive bool `uri:"active"`
			IsAdmin  bool `uri:"admin"`
		}

		params := map[string]string{
			"active": "true",
			"admin":  "false",
		}

		var flags Flags
		err := binding.BindURI(params, &flags)

		if err != nil {
			t.Errorf("Expected no error, got: %s", err.Error())
		}

		if !flags.IsActive {
			t.Error("Expected IsActive to be true")
		}

		if flags.IsAdmin {
			t.Error("Expected IsAdmin to be false")
		}
	})

	t.Run("CaseInsensitiveMatching", func(t *testing.T) {
		type User struct {
			UserID string `uri:"userid"`
			Name   string `uri:"USERNAME"`
		}

		params := map[string]string{
			"USERID":   "456",
			"username": "jane",
		}

		var user User
		err := binding.BindURI(params, &user)

		if err != nil {
			t.Errorf("Expected no error, got: %s", err.Error())
		}

		if user.UserID != "456" {
			t.Errorf("Expected UserID '456', got '%s'", user.UserID)
		}

		if user.Name != "jane" {
			t.Errorf("Expected Name 'jane', got '%s'", user.Name)
		}
	})

	t.Run("MissingParameters", func(t *testing.T) {
		type User struct {
			ID       string `uri:"id"`
			Name     string `uri:"name"`
			Optional string `uri:"optional"`
		}

		params := map[string]string{
			"id":   "123",
			"name": "john",
			// "optional" is missing
		}

		var user User
		err := binding.BindURI(params, &user)

		if err != nil {
			t.Errorf("Expected no error, got: %s", err.Error())
		}

		if user.ID != "123" {
			t.Errorf("Expected ID '123', got '%s'", user.ID)
		}

		if user.Name != "john" {
			t.Errorf("Expected Name 'john', got '%s'", user.Name)
		}

		if user.Optional != "" {
			t.Errorf("Expected Optional to be empty, got '%s'", user.Optional)
		}
	})

	t.Run("FieldsWithoutURITag", func(t *testing.T) {
		type Mixed struct {
			WithTag    string `uri:"with_tag"`
			WithoutTag string // No URI tag
		}

		params := map[string]string{
			"with_tag":    "value1",
			"without_tag": "value2",
		}

		var mixed Mixed
		err := binding.BindURI(params, &mixed)

		if err != nil {
			t.Errorf("Expected no error, got: %s", err.Error())
		}

		if mixed.WithTag != "value1" {
			t.Errorf("Expected WithTag 'value1', got '%s'", mixed.WithTag)
		}

		if mixed.WithoutTag != "" {
			t.Errorf("Expected WithoutTag to be empty, got '%s'", mixed.WithoutTag)
		}
	})

	t.Run("PointerTypes", func(t *testing.T) {
		type User struct {
			ID   *string `uri:"id"`
			Age  *int    `uri:"age"`
			Name *string `uri:"name"`
		}

		params := map[string]string{
			"id":  "123",
			"age": "25",
			// "name" is missing
		}

		var user User
		err := binding.BindURI(params, &user)

		if err != nil {
			t.Errorf("Expected no error, got: %s", err.Error())
		}

		if user.ID == nil || *user.ID != "123" {
			t.Errorf("Expected ID pointer to '123', got %v", user.ID)
		}

		if user.Age == nil || *user.Age != 25 {
			t.Errorf("Expected Age pointer to 25, got %v", user.Age)
		}

		if user.Name != nil {
			t.Errorf("Expected Name pointer to be nil, got %v", user.Name)
		}
	})

	t.Run("TimeType", func(t *testing.T) {
		type Event struct {
			CreatedAt time.Time `uri:"created_at"`
		}

		params := map[string]string{
			"created_at": "2024-01-15T10:30:00Z",
		}

		var event Event
		err := binding.BindURI(params, &event)

		if err != nil {
			t.Errorf("Expected no error, got: %s", err.Error())
		}

		expectedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
		if !event.CreatedAt.Equal(expectedTime) {
			t.Errorf("Expected CreatedAt %v, got %v", expectedTime, event.CreatedAt)
		}
	})

	t.Run("InvalidIntegerValues", func(t *testing.T) {
		type Numbers struct {
			Value int `uri:"value"`
		}

		params := map[string]string{
			"value": "not_a_number",
		}

		var numbers Numbers
		err := binding.BindURI(params, &numbers)

		// Should not return error but field should remain zero value
		if err != nil {
			t.Errorf("Expected no error, got: %s", err.Error())
		}

		if numbers.Value != 0 {
			t.Errorf("Expected Value to be 0 for invalid input, got %d", numbers.Value)
		}
	})

	t.Run("InvalidBooleanValues", func(t *testing.T) {
		type Flags struct {
			Active bool `uri:"active"`
		}

		params := map[string]string{
			"active": "maybe",
		}

		var flags Flags
		err := binding.BindURI(params, &flags)

		// Should not return error but field should remain zero value
		if err != nil {
			t.Errorf("Expected no error, got: %s", err.Error())
		}

		if flags.Active != false {
			t.Errorf("Expected Active to be false for invalid input, got %t", flags.Active)
		}
	})

	t.Run("NonPointerInput", func(t *testing.T) {
		type User struct {
			ID string `uri:"id"`
		}

		params := map[string]string{
			"id": "123",
		}

		var user User
		err := binding.BindURI(params, user) // Not a pointer

		if err == nil {
			t.Error("Expected error for non-pointer input")
		}

		if err.Error() != "v must be a pointer" {
			t.Errorf("Expected 'v must be a pointer' error, got: %s", err.Error())
		}
	})

	t.Run("NilPointer", func(t *testing.T) {
		err := binding.BindURI(map[string]string{}, nil)

		if err == nil {
			t.Error("Expected error for nil pointer")
		}

		if err.Error() != "v must be a pointer" {
			t.Errorf("Expected 'v must be a pointer' error, got: %s", err.Error())
		}
	})

	t.Run("NonStructPointer", func(t *testing.T) {
		var str string
		err := binding.BindURI(map[string]string{}, &str)

		if err == nil {
			t.Error("Expected error for non-struct pointer")
		}

		if err.Error() != "v must be a struct pointer" {
			t.Errorf("Expected 'v must be a struct pointer' error, got: %s", err.Error())
		}
	})

	t.Run("UnexportedFields", func(t *testing.T) {
		type User struct {
			ID    string `uri:"id"`
			name  string `uri:"name"` // unexported field
			Email string `uri:"email"`
		}

		params := map[string]string{
			"id":    "123",
			"name":  "john",
			"email": "john@example.com",
		}

		var user User
		err := binding.BindURI(params, &user)

		if err != nil {
			t.Errorf("Expected no error, got: %s", err.Error())
		}

		if user.ID != "123" {
			t.Errorf("Expected ID '123', got '%s'", user.ID)
		}

		if user.name != "" {
			t.Errorf("Expected unexported name to remain empty, got '%s'", user.name)
		}

		if user.Email != "john@example.com" {
			t.Errorf("Expected Email 'john@example.com', got '%s'", user.Email)
		}
	})
}

func TestFindParamCaseInsensitive(t *testing.T) {
	params := map[string]string{
		"UserID":   "123",
		"username": "john",
		"EMAIL":    "test@example.com",
	}

	t.Run("ExactMatch", func(t *testing.T) {
		value, found := findParamCaseInsensitive(params, "username")
		if !found {
			t.Error("Expected to find 'username'")
		}
		if value != "john" {
			t.Errorf("Expected value 'john', got '%s'", value)
		}
	})

	t.Run("CaseInsensitiveMatch", func(t *testing.T) {
		value, found := findParamCaseInsensitive(params, "userid")
		if !found {
			t.Error("Expected to find 'userid' (case insensitive)")
		}
		if value != "123" {
			t.Errorf("Expected value '123', got '%s'", value)
		}
	})

	t.Run("UpperCaseMatch", func(t *testing.T) {
		value, found := findParamCaseInsensitive(params, "email")
		if !found {
			t.Error("Expected to find 'email' (case insensitive)")
		}
		if value != "test@example.com" {
			t.Errorf("Expected value 'test@example.com', got '%s'", value)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		_, found := findParamCaseInsensitive(params, "nonexistent")
		if found {
			t.Error("Expected not to find 'nonexistent'")
		}
	})
}

func TestSetFieldValue(t *testing.T) {
	t.Run("StringField", func(t *testing.T) {
		var str string
		field := reflect.ValueOf(&str).Elem()

		err := setFieldValue(field, "test")
		if err != nil {
			t.Errorf("Expected no error, got: %s", err.Error())
		}

		if str != "test" {
			t.Errorf("Expected 'test', got '%s'", str)
		}
	})

	t.Run("IntField", func(t *testing.T) {
		var num int
		field := reflect.ValueOf(&num).Elem()

		err := setFieldValue(field, "42")
		if err != nil {
			t.Errorf("Expected no error, got: %s", err.Error())
		}

		if num != 42 {
			t.Errorf("Expected 42, got %d", num)
		}
	})

	t.Run("BoolField", func(t *testing.T) {
		var flag bool
		field := reflect.ValueOf(&flag).Elem()

		err := setFieldValue(field, "true")
		if err != nil {
			t.Errorf("Expected no error, got: %s", err.Error())
		}

		if !flag {
			t.Error("Expected true, got false")
		}
	})

	t.Run("Float64Field", func(t *testing.T) {
		var num float64
		field := reflect.ValueOf(&num).Elem()

		err := setFieldValue(field, "3.14")
		if err != nil {
			t.Errorf("Expected no error, got: %s", err.Error())
		}

		if num != 3.14 {
			t.Errorf("Expected 3.14, got %f", num)
		}
	})

	t.Run("PointerField", func(t *testing.T) {
		var ptr *string
		field := reflect.ValueOf(&ptr).Elem()

		err := setFieldValue(field, "test")
		if err != nil {
			t.Errorf("Expected no error, got: %s", err.Error())
		}

		if ptr == nil || *ptr != "test" {
			t.Errorf("Expected pointer to 'test', got %v", ptr)
		}
	})
}

// Benchmark tests
func BenchmarkURIBinding_BindURI(b *testing.B) {
	type User struct {
		ID    string `uri:"id"`
		Name  string `uri:"name"`
		Age   int    `uri:"age"`
		Email string `uri:"email"`
	}

	params := map[string]string{
		"id":    "123",
		"name":  "john",
		"age":   "25",
		"email": "john@example.com",
	}

	binding := URIBinding{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var user User
		binding.BindURI(params, &user)
	}
}

func BenchmarkFindParamCaseInsensitive(b *testing.B) {
	params := map[string]string{
		"UserID":   "123",
		"username": "john",
		"EMAIL":    "test@example.com",
		"age":      "25",
		"country":  "USA",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		findParamCaseInsensitive(params, "userid")
	}
}
