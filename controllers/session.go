package controllers

import "github.com/gin-gonic/gin"

func sessionUserID(c *gin.Context) uint {
	switch id := c.MustGet("session_user_id").(type) {
	case uint:    return id
	case int:     return uint(id)
	case int64:   return uint(id)
	case float64: return uint(id)
	default:      return 0
	}
}