package render

import (
	"html/template"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHTMLProduction_Render(t *testing.T) {

	t.Run("Template_nil", func(t *testing.T) {
		w := httptest.NewRecorder()

		h := HTMLProduction{
			Template: nil,
			Name:     "",
			Data:     nil,
			FuncMap:  nil,
		}

		h.Render(w)
	})

	t.Run("Valid_Template", func(t *testing.T) {
		w := httptest.NewRecorder()
		templateContent := `<h1> {{ test .Name}} </h1>`
		customFuncs := template.FuncMap{
			"test": func(titleArg string) string {
				return "Tested t: " + titleArg
			},
		}

		tmpl := template.Must(template.New("test").Funcs(customFuncs).Parse(templateContent))
		h := HTMLProduction{
			Template: tmpl,
			Name:     "",
			FuncMap:  customFuncs,
			Data: map[string]any{
				"Name": "World",
			},
		}

		err := h.Render(w)
		if err != nil {
			t.Errorf("Expected no error, got %s", err.Error())
		}

		if w.Code != 200 {
			t.Errorf("Expected status code 200, got %d", w.Code)
		}

		expectedBody := `<h1> Tested t: World </h1>`
		actualBody := w.Body.String()
		if actualBody != expectedBody {
			t.Errorf("Expected body '%s', got '%s'", expectedBody, actualBody)
		}
		contentType := w.Header().Get("Content-Type")
		expectedContentType := "text/html; charset=utf-8"
		if contentType != expectedContentType {
			t.Errorf("Expected content type '%s', got '%s'", expectedContentType, contentType)
		}
	})
}

func TestHTMLDevelopment_Render(t *testing.T) {
	t.Run("Template_nil", func(t *testing.T) {
		w := httptest.NewRecorder()

		h := HTMLDevelopment{
			TemplatePath: "",
			TemplateName: "",
			Data:         nil,
		}

		err := h.Render(w)

		expectedError := "template path is empty"

		if err.Error() != expectedError {
			t.Errorf("Expected error %s, got %s", expectedError, err.Error())
		}
	})

	t.Run("Valid_Template_File", func(t *testing.T) {
		tempDir := t.TempDir()
		templateContent := `<h1> Hello {{.Name}} </h1>`
		templatePath := filepath.Join(tempDir, "index.html")

		err := os.WriteFile(templatePath, []byte(templateContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write template file: %s", err.Error())
		}

		defer os.Remove(templatePath)

		w := httptest.NewRecorder()

		h := HTMLDevelopment{
			TemplatePath: templatePath,
			TemplateName: "",
			FuncMap:      template.FuncMap{},
			Data: map[string]any{
				"Name": "World",
			},
		}

		err = h.Render(w)

		if err != nil {
			t.Errorf("Expected no error, got %s", err.Error())
		}

		if w.Code != 200 {
			t.Errorf("Expected status code 200, got %d", w.Code)
		}

		expectedBody := `<h1> Hello World </h1>`
		actualBody := strings.TrimSpace(w.Body.String())
		if actualBody != expectedBody {
			t.Errorf("Expected body '%s', got '%s'", expectedBody, actualBody)
		}
		contentType := w.Header().Get("Content-Type")
		expectedContentType := "text/html; charset=utf-8"
		if contentType != expectedContentType {
			t.Errorf("Expected content type '%s', got '%s'", expectedContentType, contentType)
		}
	})

	t.Run("Valid_Template_FuncMap", func(t *testing.T) {
		w := httptest.NewRecorder()
		templateContent := `<h1> {{ test .Name}} </h1>`
		customFuncs := template.FuncMap{
			"test": func(titleArg string) string {
				return "Tested t: " + titleArg
			},
		}

		tempDir := t.TempDir()
		templatePath := filepath.Join(tempDir, "index.html")
		err := os.WriteFile(templatePath, []byte(templateContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write template file: %s", err.Error())
		}

		defer os.Remove(templatePath)
		h := HTMLDevelopment{
			TemplatePath: templatePath,
			TemplateName: "",
			FuncMap:      customFuncs,
			Data: map[string]any{
				"Name": "World",
			},
		}

		err = h.Render(w)
		if err != nil {
			t.Errorf("Expected no error, got %s", err.Error())
		}
		if w.Code != 200 {
			t.Errorf("Expected status code 200, got %d", w.Code)
		}
		expectedBody := `<h1> Tested t: World </h1>`
		actualBody := strings.TrimSpace(w.Body.String())
		if actualBody != expectedBody {
			t.Errorf("Expected body '%s', got '%s'", expectedBody, actualBody)
		}
		contentType := w.Header().Get("Content-Type")
		expectedContentType := "text/html; charset=utf-8"
		if contentType != expectedContentType {
			t.Errorf("Expected content type '%s', got '%s'", expectedContentType, contentType)
		}
	})
}
