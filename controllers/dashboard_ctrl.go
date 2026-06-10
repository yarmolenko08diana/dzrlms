package controllers

import (
	"net/http"

	"lms/models"
	"lms/repositories"
	"lms/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DashboardController struct {
	users   repositories.UserRepository
	courses repositories.CourseRepository
	tests   repositories.TestRepository
	assns   repositories.AssignmentRepository
	notif   *services.NotificationService
	db      *gorm.DB
}

func NewDashboardController(
	users repositories.UserRepository,
	courses repositories.CourseRepository,
	tests repositories.TestRepository,
	assns repositories.AssignmentRepository,
	notif *services.NotificationService,
	db *gorm.DB,
) *DashboardController {
	return &DashboardController{users: users, courses: courses, tests: tests, assns: assns, notif: notif, db: db}
}

func (d *DashboardController) Admin(c *gin.Context) {
	var stats models.DashboardStats
	d.db.Model(&models.User{}).Where("role = ?", models.RoleEmployee).Count(&stats.TotalEmployees)
	d.db.Model(&models.Course{}).Count(&stats.TotalCourses)
	d.db.Model(&models.Test{}).Count(&stats.TotalTests)
	d.db.Model(&models.Assignment{}).
		Where("target_type = ? AND status IN ?", models.AssignmentTypeCourse, []string{models.StatusNotStarted, models.StatusInProgress}).
		Count(&stats.ActiveCourseAssignments)
	d.db.Model(&models.Assignment{}).
		Where("target_type = ? AND status IN ?", models.AssignmentTypeTest, []string{models.StatusNotStarted, models.StatusInProgress}).
		Count(&stats.ActiveTestAssignments)
	d.db.Model(&models.Assignment{}).
		Where("status = ?", models.StatusCompleted).
		Count(&stats.CompletedAssignments)

	var recentCourse []models.Assignment
	d.db.Preload("User").Where("target_type = ?", models.AssignmentTypeCourse).
		Order("updated_at desc").Limit(5).Find(&recentCourse)

	var completedCourses []models.Assignment
	d.db.Preload("User").
		Where("target_type = ? AND status = ?", models.AssignmentTypeCourse, models.StatusCompleted).
		Order("updated_at desc").Limit(8).Find(&completedCourses)

	var completedTests []models.Assignment
	d.db.Preload("User").
		Where("target_type = ? AND status = ?", models.AssignmentTypeTest, models.StatusCompleted).
		Order("updated_at desc").Limit(8).Find(&completedTests)

	courses, _ := d.courses.FindAll()
	tests, _ := d.tests.FindAll()
	employees, _ := d.users.FindEmployees()

	c.HTML(http.StatusOK, "admin/dashboard.html", gin.H{
		"title":            "Дашборд",
		"active":           "dashboard",
		"userName":         c.MustGet("session_user_name"),
		"stats":            stats,
		"recentCourse":     recentCourse,
		"completedCourses": completedCourses,
		"completedTests":   completedTests,
		"courses":          courses,
		"tests":            tests,
		"employees":        employees,
	})
}

func (d *DashboardController) Employee(c *gin.Context) {
	userID := sessionUserID(c)
	myAssns, _ := d.assns.FindByUser(userID)

	var notStarted, inProgress, completed int
	for _, a := range myAssns {
		switch a.Status {
		case models.StatusNotStarted:
			notStarted++
		case models.StatusInProgress:
			inProgress++
		case models.StatusCompleted:
			completed++
		}
	}

	notifs, _ := d.notif.ForUser(userID)

	c.HTML(http.StatusOK, "employee/dashboard.html", gin.H{
		"title":       "Моё обучение",
		"userName":    c.MustGet("session_user_name"),
		"assignments": myAssns,
		"notStarted":  notStarted,
		"inProgress":  inProgress,
		"completed":   completed,
		"notifs":      notifs,
	})
}
