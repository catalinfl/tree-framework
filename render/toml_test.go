package render

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTOML_Render(t *testing.T) {
	t.Run("Simple_Data", func(t *testing.T) {
		w := httptest.NewRecorder()

		data := map[string]any{
			"name":   "John Doe",
			"age":    26,
			"active": true,
			"score":  12.3,
		}

		tomlRenderer := &TOML{Data: data}
		err := tomlRenderer.Render(w)
		if err != nil {
			t.Errorf("Expected no error, got %s", err.Error())
		}

		if w.Code != 200 {
			t.Errorf("Expected status code 200, got %d", w.Code)
		}

		contentType := w.Header().Get("Content-Type")
		expectedContentType := "application/toml; charset=utf-8"
		if contentType != expectedContentType {
			t.Errorf("Expected content type '%s', got '%s'", expectedContentType, contentType)
		}

		actualBody := w.Body.String()
		t.Log("Response string:", actualBody)
	})

	t.Run("Nested_Data", func(t *testing.T) {
		w := httptest.NewRecorder()

		data := map[string]interface{}{
			"title": "My App",
			"database": map[string]interface{}{
				"host":     "localhost",
				"port":     5432,
				"username": "admin",
				"ssl":      true,
			},
			"servers": []map[string]interface{}{
				{"name": "web1", "ip": "192.168.1.1"},
				{"name": "web2", "ip": "192.168.1.2"},
			},
		}

		tomlRenderer := &TOML{Data: data}
		err := tomlRenderer.Render(w)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		body := w.Body.String()

		if !strings.Contains(body, "[database]") {
			t.Errorf("Expected body to contain '[database]' table, got '%s'", body)
		}

		if !strings.Contains(body, "[[servers]]") {
			t.Errorf("Expected body to contain '[[servers]]' array, got '%s'", body)
		}

		expectedParts := []string{
			`title = "My App"`,
			`host = "localhost"`,
			`port = 5432`,
			`name = "web1"`,
			`ip = "192.168.1.1"`,
		}

		t.Log("Response string:", body)

		for _, expected := range expectedParts {
			if !strings.Contains(body, expected) {
				t.Errorf("Expected body to contain '%s', got '%s'", expected, body)
			}
		}
	})
}
