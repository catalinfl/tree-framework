package tree

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestInitMux(t *testing.T) {
	mux := InitMux()

	if mux == nil {
		t.Fatal("InitMux() returned nil")
	}

	if mux.handlers == nil {
		t.Error("handlers map not initialized")
	}

	if mux.middlewares == nil {
		t.Error("middlewares slice not initialized")
	}

	if mux.handlerMap == nil {
		t.Error("handlerMap not initialized")
	}
}

func TestMux_GET(t *testing.T) {
	mux := InitMux()

	handler := func(c *Ctx) error {
		return c.SendString("GET response", 200)
	}

	mux.GET("/test", handler)

	if routes, exists := mux.handlers[GET]; !exists {
		t.Error("GET routes not initialized")
	} else {
		if len(*routes) != 1 {
			t.Errorf("Expected 1 GET route, got %d", len(*routes))
		}

		route := (*routes)[0]
		if route.path != "/test" {
			t.Errorf("Expected path '/test', got '%s'", route.path)
		}

		if route.CtxHandler == nil {
			t.Error("CtxHandler not set")
		}
	}
}

func TestMux_HTTPMethods(t *testing.T) {
	mux := InitMux()

	handler := func(c *Ctx) error {
		return c.SendString("Response", 200)
	}

	mux.GET("/get", handler)
	mux.POST("/post", handler)
	mux.PUT("/put", handler)
	mux.DELETE("/delete", handler)
	mux.PATCH("/patch", handler)
	mux.HEAD("/head", handler)
	mux.OPTIONS("/options", handler)

	methods := []Method{GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS}
	paths := []string{"/get", "/post", "/put", "/delete", "/patch", "/head", "/options"}

	for i, method := range methods {
		if routes, exists := mux.handlers[method]; !exists {
			t.Errorf("Routes for method %s not found", method)
		} else {
			if len(*routes) != 1 {
				t.Errorf("Expected 1 route for %s, got %d", method, len(*routes))
			}

			route := (*routes)[0]
			if route.path != paths[i] {
				t.Errorf("Expected path '%s' for %s, got '%s'", paths[i], method, route.path)
			}
		}
	}
}

func TestMux_ParameterRoutes(t *testing.T) {
	mux := InitMux()

	handler := func(c *Ctx) error {
		id, err := c.GetURLParam("id")
		if err != nil {
			return c.SendString("Parameter not found", 400)
		}
		return c.SendString("ID: "+id, 200)
	}

	mux.GET("/users/:id", handler)

	req := httptest.NewRequest("GET", "/users/123", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "ID: 123" {
		t.Errorf("Expected 'ID: 123', got '%s'", string(body))
	}
}

func TestMux_MultipleParameters(t *testing.T) {
	mux := InitMux()

	handler := func(c *Ctx) error {
		userID, _ := c.GetURLParam("userId")
		postID, _ := c.GetURLParam("postId")
		response := fmt.Sprintf("User: %s, Post: %s", userID, postID)
		return c.SendString(response, 200)
	}

	mux.GET("/users/:userId/posts/:postId", handler)

	req := httptest.NewRequest("GET", "/users/john/posts/456", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	expected := "User: john, Post: 456"

	if string(body) != expected {
		t.Errorf("Expected '%s', got '%s'", expected, string(body))
	}
}

func TestMux_Middleware(t *testing.T) {
	mux := InitMux()

	middleware := func(c *Ctx) error {
		c.w.Header().Set("X-Test-Middleware", "applied")
		return nil
	}

	mux.USE("/", middleware)

	handler := func(c *Ctx) error {
		return c.SendString("OK", 200)
	}

	mux.GET("/test", handler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	resp := w.Result()
	if resp.Header.Get("X-Test-Middleware") != "applied" {
		t.Error("Middleware was not applied")
	}
}

func TestMux_JSONBinding(t *testing.T) {
	mux := InitMux()
	type User struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	handler := func(c *Ctx) error {
		var user User
		if err := c.BindJSON(&user); err != nil {
			return c.SendString("Binding failed", 400)
		}

		response := J{
			"received_name":  user.Name,
			"received_email": user.Email,
		}
		return c.SendJSON(response, 200)
	}
	mux.POST("/users", handler)

	userData := User{
		Name:  "John Doe",
		Email: "john@example.com",
	}
	jsonData, _ := json.Marshal(userData)
	req := httptest.NewRequest("POST", "/users", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	resp := w.Result()
	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
	var response map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&response)
	if response["received_name"] != "John Doe" {
		t.Errorf("Expected name 'John Doe', got %v", response["received_name"])
	}
}

func TestMux_FormBinding(t *testing.T) {
	mux := InitMux()

	handler := func(c *Ctx) error {
		name, err := c.FormValue("name")
		if err != nil {
			return c.SendString("Name required", 400)
		}

		email, err := c.FormValue("email")
		if err != nil {
			return c.SendString("Email required", 400)
		}

		response := fmt.Sprintf("Name: %s, Email: %s", name, email)
		return c.SendString(response, 200)
	}

	mux.POST("/form", handler)

	form := url.Values{}
	form.Add("name", "Jane Doe")
	form.Add("email", "jane@example.com")

	req := httptest.NewRequest("POST", "/form", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	expected := "Name: Jane Doe, Email: jane@example.com"

	if string(body) != expected {
		t.Errorf("Expected '%s', got '%s'", expected, string(body))
	}
}

func TestMux_FileUpload(t *testing.T) {
	mux := InitMux()

	handler := func(c *Ctx) error {
		file, err := c.FormFile("upload")
		if err != nil {
			return c.SendString("File upload failed", 400)
		}

		response := J{
			"filename": file.MultipartFileHeader.Filename,
			"size":     file.MultipartFileHeader.Size,
		}
		return c.SendJSON(response, 200)
	}

	mux.POST("/upload", handler)

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	fileWriter, _ := writer.CreateFormFile("upload", "test.txt")
	fileWriter.Write([]byte("test file content"))
	writer.Close()

	req := httptest.NewRequest("POST", "/upload", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var response map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&response)

	if response["filename"] != "test.txt" {
		t.Errorf("Expected filename 'test.txt', got %v", response["filename"])
	}
}

func TestMux_QueryParameters(t *testing.T) {
	mux := InitMux()

	handler := func(c *Ctx) error {
		search, err := c.GetQuery("q")
		if err != nil {
			return c.SendString("Query parameter missing", 400)
		}

		page, err := c.GetQueryInt("page")
		if err != nil {
			page = 1
		}

		response := J{
			"search": search,
			"page":   page,
		}
		return c.SendJSON(response, 200)
	}

	mux.GET("/search", handler)

	req := httptest.NewRequest("GET", "/search?q=golang&page=2", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	resp := w.Result()
	var response map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&response)

	if response["search"] != "golang" {
		t.Errorf("Expected search 'golang', got %v", response["search"])
	}

	if response["page"] != float64(2) {
		t.Errorf("Expected page 2, got %v", response["page"])
	}
}

func TestMux_Headers(t *testing.T) {
	mux := InitMux()

	handler := func(c *Ctx) error {
		userAgent, err := c.GetHeader("User-Agent")
		if err != nil {
			return c.SendString("User-Agent header missing", 400)
		}

		c.SetHeader("X-Custom-Header", "custom-value")

		return c.SendString("User-Agent: "+userAgent, 200)
	}

	mux.GET("/headers", handler)

	req := httptest.NewRequest("GET", "/headers", nil)
	req.Header.Set("User-Agent", "Test-Agent/1.0")

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if string(body) != "User-Agent: Test-Agent/1.0" {
		t.Errorf("Expected 'User-Agent: Test-Agent/1.0', got '%s'", string(body))
	}

	if resp.Header.Get("X-Custom-Header") != "custom-value" {
		t.Error("Custom header not set")
	}
}

func TestMux_Cookies(t *testing.T) {
	mux := InitMux()

	handler := func(c *Ctx) error {

		sessionID, err := c.GetCookieValue("session_id")
		if err != nil {
			sessionID = "new-session"
		}

		cookie := &Cookie{
			Name:     "user_pref",
			Value:    "dark_mode",
			MaxAge:   3600,
			HttpOnly: true,
		}
		c.SetCookie(cookie)

		return c.SendString("Session: "+sessionID, 200)
	}

	mux.GET("/profile", handler)

	req := httptest.NewRequest("GET", "/profile", nil)
	req.AddCookie(&http.Cookie{
		Name:  "session_id",
		Value: "abc123",
	})

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if string(body) != "Session: abc123" {
		t.Errorf("Expected 'Session: abc123', got '%s'", string(body))
	}

	cookies := resp.Cookies()
	found := false
	for _, cookie := range cookies {
		if cookie.Name == "user_pref" && cookie.Value == "dark_mode" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Cookie 'user_pref' not set correctly")
	}
}

func TestMux_NotFound(t *testing.T) {
	mux := InitMux()

	handler := func(c *Ctx) error {
		return c.SendString("Found", 200)
	}

	mux.GET("/exists", handler)

	req := httptest.NewRequest("GET", "/does-not-exist", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != 404 {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}
}

func TestMux_RouteConflicts(t *testing.T) {
	mux := InitMux()

	handler1 := func(c *Ctx) error {
		return c.SendString("Static route", 200)
	}

	handler2 := func(c *Ctx) error {
		param, _ := c.GetURLParam("id")
		return c.SendString("Param route: "+param, 200)
	}

	mux.GET("/users/profile", handler1)
	mux.GET("/users/:id", handler2)

	req1 := httptest.NewRequest("GET", "/users/profile", nil)
	w1 := httptest.NewRecorder()
	mux.ServeHTTP(w1, req1)

	body1, _ := io.ReadAll(w1.Result().Body)
	if string(body1) != "Static route" {
		t.Errorf("Expected 'Static route', got '%s'", string(body1))
	}

	req2 := httptest.NewRequest("GET", "/users/123", nil)
	w2 := httptest.NewRecorder()
	mux.ServeHTTP(w2, req2)

	body2, _ := io.ReadAll(w2.Result().Body)
	if string(body2) != "Param route: 123" {
		t.Errorf("Expected 'Param route: 123', got '%s'", string(body2))
	}
}

func BenchmarkMux_SimpleRoute(b *testing.B) {
	mux := InitMux()

	handler := func(c *Ctx) error {
		return c.SendString("OK", 200)
	}

	mux.GET("/test", handler)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
	}
}

func BenchmarkMux_ParameterRoute(b *testing.B) {
	mux := InitMux()

	handler := func(c *Ctx) error {
		id, _ := c.GetURLParam("id")
		return c.SendString("ID: "+id, 200)
	}

	mux.GET("/users/:id", handler)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/users/123", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
	}
}

func TestMux_ConcurrentRequests(t *testing.T) {
	mux := InitMux()

	handler := func(c *Ctx) error {

		time.Sleep(10 * time.Millisecond)
		return c.SendString("OK", 200)
	}

	mux.GET("/concurrent", handler)

	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			req := httptest.NewRequest("GET", "/concurrent", nil)
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)

			if w.Result().StatusCode != 200 {
				t.Error("Concurrent request failed")
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}
