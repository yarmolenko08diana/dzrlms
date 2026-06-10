package controllers

import (
	"net/http"
	"lms/models"
	"lms/repositories"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type AuthController struct {
	users repositories.UserRepository
}

func NewAuthController(users repositories.UserRepository) *AuthController {
	return &AuthController{users: users}
}

func (a *AuthController) LoginPage(c *gin.Context) {
	session := sessions.Default(c)
	if session.Get("user_id") != nil {
		if session.Get("user_role") == models.RoleAdmin {
			c.Redirect(http.StatusFound, "/admin/dashboard")
		} else {
			c.Redirect(http.StatusFound, "/employee/dashboard")
		}
		return
	}
	c.HTML(http.StatusOK, "auth/login.html", gin.H{"title": "Вход"})
}

func (a *AuthController) Login(c *gin.Context) {
	email    := c.PostForm("email")
	password := c.PostForm("password")

	user, err := a.users.FindByEmail(email)
	if err != nil {
		c.HTML(http.StatusUnauthorized, "auth/login.html", gin.H{"error": "Неверный email или пароль."})
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
		c.HTML(http.StatusUnauthorized, "auth/login.html", gin.H{"error": "Неверный email или пароль."})
		return
	}

	session := sessions.Default(c)
	session.Set("user_id", user.ID)
	session.Set("user_name", user.Name)
	session.Set("user_role", user.Role)
	session.Save()

	if user.Role == models.RoleAdmin {
		c.Redirect(http.StatusFound, "/admin/dashboard")
	} else {
		c.Redirect(http.StatusFound, "/employee/dashboard")
	}
}

func (a *AuthController) Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()
	c.Redirect(http.StatusFound, "/login")
}