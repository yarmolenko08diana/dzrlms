package middleware

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userID := session.Get("user_id")
		if userID == nil {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}
		c.Next()
	}
}

func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		role := session.Get("user_role")
		if role != "admin" {
			c.Redirect(http.StatusFound, "/employee/dashboard")
			c.Abort()
			return
		}
		c.Next()
	}
}

func InjectUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		c.Set("session_user_id", session.Get("user_id"))
		c.Set("session_user_name", session.Get("user_name"))
		c.Set("session_user_role", session.Get("user_role"))
		c.Next()
	}
}