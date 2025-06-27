package tree

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func BenchmarkRouting_Static(b *testing.B) {
	mux := InitMux()

	handler := func(c *Ctx) error {
		return c.SendString("OK", 200)
	}

	mux.GET("/", handler)
	mux.GET("/users", handler)
	mux.GET("/api/v1/users", handler)
	mux.GET("/api/v1/posts", handler)
	mux.GET("/api/v2/comments", handler)

	req := httptest.NewRequest("GET", "/api/v1/users", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
	}
}

func BenchmarkRouting_Parameters(b *testing.B) {
	mux := InitMux()

	handler := func(c *Ctx) error {
		id, _ := c.GetURLParam("id")
		return c.SendString("User: "+id, 200)
	}

	mux.GET("/users/:id", handler)
	mux.GET("/users/:id/posts/:postId", handler)
	mux.GET("/api/v1/users/:userId/comments/:commentId", handler)

	req := httptest.NewRequest("GET", "/users/123", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
	}
}

func BenchmarkRouting_Regex(b *testing.B) {
	mux := InitMux()

	handler := func(c *Ctx) error {
		param, _ := c.RegexURLParam(1)
		return c.SendString("Regex: "+param, 200)
	}

	mux.GET("/users/:|\\d+|", handler)
	mux.GET("/files/:|[a-zA-Z0-9]+|/download", handler)

	req := httptest.NewRequest("GET", "/users/123", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
	}
}

type UserForm struct {
	ID     int    `json:"id" form:"id"`
	Name   string `json:"name" form:"name"`
	Email  string `json:"email" form:"email"`
	Age    int    `json:"age" form:"age"`
	Active bool   `json:"active" form:"active"`
}

func BenchmarkBinding_JSON(b *testing.B) {
	mux := InitMux()

	handler := func(c *Ctx) error {
		var user UserForm
		if err := c.BindJSON(&user); err != nil {
			return c.SendString("Error", 400)
		}
		return c.SendJSON(J{"test": user}, 200)
	}

	mux.POST("/user", handler)

	userData := UserForm{
		ID:     123,
		Name:   "John Doe",
		Email:  "john@example.com",
		Age:    30,
		Active: true,
	}

	jsonData, _ := json.Marshal(userData)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/user", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
	}
}

func BenchmarkRouting_Mixed(b *testing.B) {
	mux := InitMux()

	handler := func(c *Ctx) error {
		return c.SendString("OK", 200)
	}

	// Add many routes to simulate real application
	routes := []string{
		"/",
		"/about",
		"/contact",
		"/users",
		"/users/:id",
		"/users/:id/posts",
		"/users/:id/posts/:postId",
		"/api/v1/users",
		"/api/v1/users/:id",
		"/api/v1/posts",
		"/api/v2/comments",
		"/admin/dashboard",
		"/admin/users",
		"/admin/settings",
		"/files/:|\\d+|/download",
		"/products/:|[A-Z]{2}\\d{4}|",
	}

	for _, route := range routes {
		mux.GET(route, handler)
	}

	// Test requests
	requests := []*http.Request{
		httptest.NewRequest("GET", "/", nil),
		httptest.NewRequest("GET", "/users/123", nil),
		httptest.NewRequest("GET", "/api/v1/users/456", nil),
		httptest.NewRequest("GET", "/files/789/download", nil),
		httptest.NewRequest("GET", "/products/AB1234", nil),
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := requests[i%len(requests)]
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
	}
}
