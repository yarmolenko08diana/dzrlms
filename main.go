package main

import (
	"html/template"
	"log"
	"os"
	"strings"
	"time"

	"lms/db"
	"lms/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	db.Connect()
	db.Migrate()
	db.Seed()

	r := gin.Default()

	r.SetFuncMap(template.FuncMap{
		"safe":        func(s string) template.JS   { return template.JS(s) },
		"safeHTML":    func(s string) template.HTML { return template.HTML(s) },
		"safeURL":     func(s string) template.URL  { return template.URL(s) },
		"formatDate":  func(t time.Time) string     { return t.Format("02.01.2006") },
		"formatTime":  func(t time.Time) string     { return t.Format("02.01.2006 15:04") },
		"add":         func(a, b int) int { return a + b },
		"sub":         func(a, b int) int { return a - b },
		"mul":         func(a, b int) int { return a * b },
		"div":         func(a, b int) int { if b == 0 { return 0 }; return a / b },
		"inc":         func(i int) int { return i + 1 },
		"contains":    func(s, substr string) bool { return strings.Contains(s, substr) },
		"isCourse":    func(s string) bool { return s == "course" },
		"isTest":      func(s string) bool { return s == "test" },
	})

	r.LoadHTMLGlob("templates/**/*.html")
	r.Static("/static", "./static")
	r.Static("/uploads", "./static/uploads")

	routes.Setup(r, db.DB)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("LMS запущен: http://localhost:%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
