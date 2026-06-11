package main

import (
	"html/template"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"lms/db"
	"lms/routes"

	"github.com/gin-gonic/gin"
)

var youtubeRe = regexp.MustCompile(`(?:youtube\.com/watch\?v=|youtu\.be/)([a-zA-Z0-9_-]{11})`)

func youtubeEmbedURL(url string) string {
	m := youtubeRe.FindStringSubmatch(url)
	if m == nil {
		return ""
	}
	return "https://www.youtube.com/embed/" + m[1]
}

var videoExts = []string{".mp4", ".webm", ".ogg", ".mov", ".m4v"}
var imageExts = []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg"}

func isVideoURL(url string) bool {
	lower := strings.ToLower(url)
	for _, ext := range videoExts {
		if strings.HasSuffix(lower, ext) {
			return true
		}
	}
	return strings.Contains(lower, "youtube.com") || strings.Contains(lower, "youtu.be")
}

func isImageURL(url string) bool {
	lower := strings.ToLower(url)
	for _, ext := range imageExts {
		if strings.HasSuffix(lower, ext) {
			return true
		}
	}
	return false
}

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
		"isVideoURL":      isVideoURL,
		"isImageURL":      isImageURL,
		"youtubeEmbedURL": youtubeEmbedURL,
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
	if err := r.Run("127.0.0.1:" + port); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}