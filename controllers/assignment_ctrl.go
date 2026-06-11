package controllers

import (
	"net/http"

	"lms/models"
	"lms/repositories"
	"lms/services"

	"github.com/gin-gonic/gin"
)

type AssignmentController struct {
	svc     *services.AssignmentService
	tests   *services.TestService
	courses *services.CourseService
	users   repositories.UserRepository
}

func NewAssignmentController(svc *services.AssignmentService, tests *services.TestService, courses *services.CourseService, users repositories.UserRepository) *AssignmentController {
	return &AssignmentController{svc: svc, tests: tests, courses: courses, users: users}
}

func (a *AssignmentController) Assign(c *gin.Context) {

	var req services.AssignBatchRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if len(req.TargetTypes) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "target_types is required",
		})
		return
	}

	if err := a.svc.AssignBatch(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok": true,
	})
}

func (a *AssignmentController) List(c *gin.Context) {

	targetType := c.Query("type")
	status := c.Query("status")

	list, err := a.svc.List(targetType, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	courses, _ := a.courses.List()
	tests, _ := a.tests.List()
	employees, _ := a.users.FindEmployees()

	courseTitles := make(map[uint]string, len(courses))
	for _, course := range courses {
		courseTitles[course.ID] = course.Title
	}
	testTitles := make(map[uint]string, len(tests))
	for _, test := range tests {
		testTitles[test.ID] = test.Title
	}

	c.HTML(http.StatusOK, "admin/assignments.html", gin.H{
		"title":        "Назначения",
		"rows":         list,
		"active":       "assignments",
		"userName":     c.MustGet("session_user_name"),
		"filterType":   targetType,
		"filterStatus": status,
		"courses":      courses,
		"tests":        tests,
		"employees":    employees,
		"courseTitles": courseTitles,
		"testTitles":   testTitles,
	})
}
func (a *AssignmentController) Results(c *gin.Context) {
	id := parseID(c, "id")

	assn, prog, answers, err := a.svc.GetTestResults(id)
	if err != nil {
		c.Redirect(http.StatusFound, "/admin/assignments")
		return
	}

	test, _ := a.tests.Get(assn.TargetID)

	answersByQ := make(map[uint]models.TestAnswer, len(answers))
	for _, ta := range answers {
		answersByQ[ta.QuestionID] = ta
	}

	c.HTML(http.StatusOK, "admin/test_results.html", gin.H{
		"title":      "Результаты теста",
		"active":     "assignments",
		"userName":   c.MustGet("session_user_name"),
		"assignment": assn,
		"progress":   prog,
		"answers":    answers,
		"answersByQ": answersByQ,
		"test":       test,
	})
}

func (a *AssignmentController) Delete(c *gin.Context) {

	id := parseID(c, "id")

	if err := a.svc.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.Redirect(http.StatusFound, "/admin/assignments")
}