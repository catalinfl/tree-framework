package render

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestJSONRenderers(t *testing.T) { // Renamed from PureJSONTest to TestPureJSON

	t.Run("PureJSON_Subtest", func(t *testing.T) { // Also good practice to make subtest names distinct if needed
		w := httptest.NewRecorder()
		pureJSON := PureJSON{
			Data: map[string]any{
				"key":  "value",
				"html": "<p>hello</p>", // Add a case that benefits from PureJSON
			},
		}

		err := pureJSON.Render(w)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code 200, got %d", w.Code)
		}

		// log.Println("Response Headers:", w.Header()) // Logging is fine for debugging but often removed from final tests

		contentType := w.Header().Get("Content-Type")
		if contentType != "application/json; charset=utf-8" {
			t.Errorf("Expected content type 'application/json; charset=utf-8', got '%s'", contentType)
		}

		expectedBody := `{"html":"<p>hello</p>","key":"value"}`

		actualBody := strings.TrimSpace(w.Body.String())

		if !((strings.Contains(actualBody, `"key":"value"`) && strings.Contains(actualBody, `"html":"<p>hello</p>"`)) && len(actualBody) == len(expectedBody)) {
			t.Errorf("Expected body to contain elements of '%s', got '%s'", expectedBody, actualBody)
		}
	})

	t.Run("JSONP_Subtest", func(t *testing.T) {
		w := httptest.NewRecorder()
		jsonp := JSONP{
			Callback: "callbackFunction",
			Data: map[string]any{
				"key":   "value",
				"fruit": "banana",
			},
		}

		err := jsonp.Render(w)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code 200, got %d", w.Code)
		}

		contentType := w.Header().Get("Content-Type")
		if contentType != "application/javascript; charset=utf-8" {
			t.Errorf("Expected content type 'application/javascript; charset=utf-8', got '%s'", contentType)
		}
		expectedBody := `callbackFunction({"fruit":"banana","key":"value"});`
		actualBody := strings.TrimSpace(w.Body.String())
		if actualBody != expectedBody {
			t.Errorf("Expected body '%s', got '%s'", expectedBody, actualBody)
		}
	})

	t.Run("ASCIIJSON_Subtest", func(t *testing.T) {
		w := httptest.NewRecorder()
		asciiJSON := ASCIIJSON{
			Data: map[string]any{
				"key":   "value",
				"fruit": "banana⚠️",
			},
		}

		err := asciiJSON.Render(w)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code 200, got %d", w.Code)
		}

		contentType := w.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected content type 'application/json', got '%s'", contentType)
		}

		t.Log("Response Headers:", w.Body.String())
	})

	t.Run("SecureJSON_Subtest", func(t *testing.T) {
		w := httptest.NewRecorder()
		secureJSON := SecureJSON{
			Prefix: ")]}'",
			Data: []map[string]any{
				{
					"key": "value",
				},
				{
					"fruit": "banana",
				},
			},
		}

		err := secureJSON.Render(w)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code 200, got %d", w.Code)
		}

		contentType := w.Header().Get("Content-Type")
		if contentType != "application/json; charset=utf-8" {
			t.Errorf("Expected content type 'application/json; charset=utf-8', got '%s'", contentType)
		}

		expectedBody := `)]}'[{"key":"value"},{"fruit":"banana"}]`
		actualBody := strings.TrimSpace(w.Body.String())
		if actualBody != expectedBody {
			t.Errorf("Expected body '%s', got '%s'", expectedBody, actualBody)
		}
	})
}
