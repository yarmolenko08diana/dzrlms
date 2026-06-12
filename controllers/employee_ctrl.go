package controllers

import (
	"log"
	"net/http"
	"strconv"

	"lms/repositories"
	"lms/services"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type EmployeeController struct {
	users repositories.UserRepository
}

func NewEmployeeController(users repositories.UserRepository) *EmployeeController {
	return &EmployeeController{users: users}
}

func (e *EmployeeController) List(c *gin.Context) {
	employees, _ := e.users.FindEmployees()
	c.HTML(http.StatusOK, "admin/employees.html", gin.H{
		"title":     "Сотрудники",
		"employees": employees,
		"active":    "employees",
		"userName":  c.MustGet("session_user_name"),
	})
}

func (e *EmployeeController) NewForm(c *gin.Context) {
	c.HTML(http.StatusOK, "admin/employee_form.html", gin.H{
		"title":    "Добавить сотрудника",
		"active":   "employees",
		"userName": c.MustGet("session_user_name"),
	})
}

func (e *EmployeeController) Create(c *gin.Context) {
	name, email, password := c.PostForm("name"), c.PostForm("email"), c.PostForm("password")
	if name == "" || email == "" || password == "" {
		c.HTML(http.StatusBadRequest, "admin/employee_form.html", gin.H{"error": "Все поля обязательны."})
		return
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	u := services.NewUserModel(name, email, string(hash))
	e.users.Create(&u)
	c.Redirect(http.StatusFound, "/admin/employees")
}

func (e *EmployeeController) EditForm(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	user, err := e.users.FindByID(uint(id))
	if err != nil {
		c.Redirect(http.StatusFound, "/admin/employees")
		return
	}
	c.HTML(http.StatusOK, "admin/employee_form.html", gin.H{
		"title":    "Редактировать сотрудника",
		"active":   "employees",
		"userName": c.MustGet("session_user_name"),
		"employee": user,
		"editing":  true,
	})
}

func (e *EmployeeController) Update(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	user, err := e.users.FindByID(uint(id))
	if err != nil {
		c.Redirect(http.StatusFound, "/admin/employees")
		return
	}
	user.Name = c.PostForm("name")
	user.Email = c.PostForm("email")
	if pw := c.PostForm("password"); pw != "" {
		hash, _ := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
		user.Password = string(hash)
	}
	e.users.Update(user)
	c.Redirect(http.StatusFound, "/admin/employees")
}

func (e *EmployeeController) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := e.users.Delete(uint(id)); err != nil {
		log.Printf("employee delete failed: %v", err)
	}
	c.Redirect(http.StatusFound, "/admin/employees")
}