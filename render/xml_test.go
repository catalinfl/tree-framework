package render

import (
	"encoding/xml"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestXML_Render(t *testing.T) {
	t.Run("Simple_Data", func(t *testing.T) {
		w := httptest.NewRecorder()

		type Person struct {
			XMLName xml.Name `xml:"person"`
			Name    string   `xml:"name"`
			Age     int      `xml:"age"`
			Active  bool     `xml:"active"`
			Score   float64  `xml:"score"`
		}

		data := Person{
			Name:   "John Doe",
			Age:    26,
			Active: true,
			Score:  12.3,
		}

		xmlRenderer := &XML{Data: data}
		err := xmlRenderer.Render(w)
		if err != nil {
			t.Errorf("Expected no error, got %s", err.Error())
		}

		if w.Code != 200 {
			t.Errorf("Expected status code 200, got %d", w.Code)
		}

		contentType := w.Header().Get("Content-Type")
		expectedContentType := "application/xml; charset=utf-8"
		if contentType != expectedContentType {
			t.Errorf("Expected content type '%s', got '%s'", expectedContentType, contentType)
		}

		actualBody := w.Body.String()

		expectedParts := []string{
			`<person>`,
			`<name>John Doe</name>`,
			`<age>26</age>`,
			`<active>true</active>`,
			`<score>12.3</score>`,
			`</person>`,
		}

		for _, expected := range expectedParts {
			if !strings.Contains(actualBody, expected) {
				t.Errorf("Expected body to contain '%s', got '%s'", expected, actualBody)
			}
		}
	})

	t.Run("Nested_Data", func(t *testing.T) {
		w := httptest.NewRecorder()

		type Database struct {
			Host     string `xml:"host"`
			Port     int    `xml:"port"`
			Username string `xml:"username"`
			SSL      bool   `xml:"ssl"`
		}

		type Server struct {
			Name string `xml:"name"`
			IP   string `xml:"ip"`
		}

		type Config struct {
			XMLName  xml.Name `xml:"config"`
			Title    string   `xml:"title"`
			Database Database `xml:"database"`
			Servers  []Server `xml:"servers>server"`
		}

		data := Config{
			Title: "My App",
			Database: Database{
				Host:     "localhost",
				Port:     5432,
				Username: "admin",
				SSL:      true,
			},
			Servers: []Server{
				{Name: "web1", IP: "192.168.1.1"},
				{Name: "web2", IP: "192.168.1.2"},
			},
		}

		xmlRenderer := &XML{Data: data}
		err := xmlRenderer.Render(w)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		body := w.Body.String()
		t.Log("Response string:", body)

		// Check for nested elements
		if !strings.Contains(body, "<database>") {
			t.Errorf("Expected body to contain '<database>' element, got '%s'", body)
		}

		if !strings.Contains(body, "<servers>") {
			t.Errorf("Expected body to contain '<servers>' element, got '%s'", body)
		}

		expectedParts := []string{
			`<config>`,
			`<title>My App</title>`,
			`<host>localhost</host>`,
			`<port>5432</port>`,
			`<username>admin</username>`,
			`<ssl>true</ssl>`,
			`<name>web1</name>`,
			`<ip>192.168.1.1</ip>`,
			`<name>web2</name>`,
			`<ip>192.168.1.2</ip>`,
			`</config>`,
		}

		t.Log("Response string:", body)

		for _, expected := range expectedParts {
			if !strings.Contains(body, expected) {
				t.Errorf("Expected body to contain '%s', got '%s'", expected, body)
			}
		}
	})
}
