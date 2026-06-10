package routes

import (
	"os"

	"lms/controllers"
	"lms/middleware"
	"lms/services"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Setup(r *gin.Engine, gormDB *gorm.DB) {
	secret := os.Getenv("SESSION_SECRET")
	if secret == "" {
		secret = "lms-secret-key-change-in-production"
	}
	store := cookie.NewStore([]byte(secret))
	r.Use(sessions.Sessions("lms_session", store))
	r.Use(middleware.InjectUser())

	svc := services.NewContainer(gormDB)
	authCtrl := controllers.NewAuthController(svc.Repos.Users)
	empCtrl := controllers.NewEmployeeController(svc.Repos.Users)
	courseCtrl := controllers.NewCourseController(svc.Courses, svc.Assignments, svc.Repos.Users, svc.Upload)
	testCtrl := controllers.NewTestController(svc.Tests, svc.Assignments, svc.Repos.Users, svc.Upload)
	assnCtrl := controllers.NewAssignmentController(svc.Assignments)
	dashCtrl := controllers.NewDashboardController(svc.Repos.Users, svc.Repos.Courses, svc.Repos.Tests, svc.Repos.Assignments, svc.Notif, gormDB)

	r.GET("/", func(c *gin.Context) { c.Redirect(302, "/login") })
	r.GET("/login", authCtrl.LoginPage)
	r.POST("/login", authCtrl.Login)
	r.GET("/logout", authCtrl.Logout)

	admin := r.Group("/admin", middleware.RequireAuth(), middleware.RequireAdmin())
	{
		admin.GET("/dashboard", dashCtrl.Admin)

		admin.GET("/employees", empCtrl.List)
		admin.GET("/employees/new", empCtrl.NewForm)
		admin.POST("/employees", empCtrl.Create)
		admin.GET("/employees/:id/edit", empCtrl.EditForm)
		admin.POST("/employees/:id", empCtrl.Update)
		admin.POST("/employees/:id/delete", empCtrl.Delete)
		admin.GET("/employees/:id/profile", empCtrl.Profile)

		admin.GET("/courses", courseCtrl.List)
		admin.GET("/courses/new", courseCtrl.NewEditor)
		admin.GET("/courses/:id/edit", courseCtrl.EditEditor)
		admin.POST("/courses/save", courseCtrl.Save)     // new
		admin.POST("/courses/:id/save", courseCtrl.Save) // edit
		admin.POST("/courses/:id/publish", courseCtrl.Publish)
		admin.POST("/courses/:id/delete", courseCtrl.Delete)
		admin.POST("/courses/upload", courseCtrl.UploadMedia)

		admin.GET("/tests", testCtrl.List)
		admin.GET("/tests/new", testCtrl.NewEditor)
		admin.GET("/tests/:id/edit", testCtrl.EditEditor)
		admin.POST("/tests/save", testCtrl.Save)
		admin.POST("/tests/:id/save", testCtrl.Save)
		admin.POST("/tests/:id/publish", testCtrl.Publish)
		admin.POST("/tests/:id/delete", testCtrl.Delete)
		admin.POST("/tests/upload", testCtrl.UploadMedia)

		admin.GET("/assignments", assnCtrl.List)
		admin.POST("/assignments/assign", assnCtrl.Assign)
		admin.POST("/assignments/:id/delete", assnCtrl.Delete)
	}

	emp := r.Group("/employee", middleware.RequireAuth())
	{
		emp.GET("/dashboard", dashCtrl.Employee)
		emp.GET("/courses/:id", courseCtrl.EmployeeView)
		emp.POST("/courses/:id/complete", courseCtrl.CompleteCourse)
		emp.GET("/tests/:id", testCtrl.EmployeeView)
		emp.POST("/tests/:id/submit", testCtrl.SubmitAnswer)
		emp.POST("/tests/:id/complete", testCtrl.CompleteTest)
	}
}
