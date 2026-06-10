package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"lms/models"
	"lms/repositories"
	"lms/services"

	"github.com/gin-gonic/gin"
)

type CourseController struct {
	svc     *services.CourseService
	assnSvc *services.AssignmentService
	users   repositories.UserRepository
	upload  *services.UploadService
}

func NewCourseController(
	svc *services.CourseService,
	assnSvc *services.AssignmentService,
	users repositories.UserRepository,
	upload *services.UploadService,
) *CourseController {
	return &CourseController{svc: svc, assnSvc: assnSvc, users: users, upload: upload}
}

func (co *CourseController) List(c *gin.Context) {
	courses, _ := co.svc.List()
	c.HTML(http.StatusOK, "admin/courses.html", gin.H{
		"title":    "Курсы",
		"courses":  courses,
		"active":   "courses",
		"userName": c.MustGet("session_user_name"),
	})
}

func (co *CourseController) NewEditor(c *gin.Context) {
	employees, _ := co.users.FindEmployees()
	c.HTML(http.StatusOK, "admin/course_editor.html", gin.H{
		"title":     "Создать курс",
		"active":    "courses",
		"userName":  c.MustGet("session_user_name"),
		"employees": employees,
	})
}

func (co *CourseController) EditEditor(c *gin.Context) {
	id := parseID(c, "id")
	course, err := co.svc.Get(id)
	if err != nil {
		c.Redirect(http.StatusFound, "/admin/courses")
		return
	}
	employees, _ := co.users.FindEmployees()
	courseJSON, _ := json.Marshal(course)
	c.HTML(http.StatusOK, "admin/course_editor.html", gin.H{
		"title":      "Редактировать курс",
		"active":     "courses",
		"userName":   c.MustGet("session_user_name"),
		"course":     course,
		"courseJSON": string(courseJSON),
		"employees":  employees,
		"editing":    true,
	})
}

func (co *CourseController) Save(c *gin.Context) {
	courseID := parseID(c, "id") // 0 for new
	var payload services.SaveCoursePayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	course, err := co.svc.Upsert(courseID, payload)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": course.ID})
}

func (co *CourseController) Publish(c *gin.Context) {
	courseID := parseID(c, "id")
	var req struct {
		AssignAll bool   `json:"assign_all"`
		UserIDs   []uint `json:"user_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := co.svc.Publish(courseID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := co.assnSvc.Assign(services.AssignRequest{
		TargetType: models.AssignmentTypeCourse,
		TargetID:   courseID,
		AssignAll:  req.AssignAll,
		UserIDs:    req.UserIDs,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (co *CourseController) Delete(c *gin.Context) {
	if err := co.svc.Delete(parseID(c, "id")); err != nil {
		log.Printf("course delete failed: %v", err)
	}
	c.Redirect(http.StatusFound, "/admin/courses")
}

func (co *CourseController) UploadMedia(c *gin.Context) {
	fh, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no file"})
		return
	}
	url, err := co.upload.Save(fh)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"url": url})
}

func (co *CourseController) EmployeeView(c *gin.Context) {
	courseID := parseID(c, "id")
	userID := sessionUserID(c)

	a, prog, err := co.assnSvc.OpenCourse(userID, courseID)
	if err != nil {
		c.Redirect(http.StatusFound, "/employee/dashboard")
		return
	}
	course, err := co.svc.Get(courseID)
	if err != nil {
		c.Redirect(http.StatusFound, "/employee/dashboard")
		return
	}

	slideIDParam := parseID(c, "slide") // from query ?slide=X
	if slideIDParam == 0 && prog.CurrentSlideID != nil {
		slideIDParam = *prog.CurrentSlideID
	}
	if slideIDParam == 0 && len(course.Slides) > 0 {
		slideIDParam = course.Slides[0].ID
	}

	currentIdx := 0
	for i, s := range course.Slides {
		if s.ID == slideIDParam {
			currentIdx = i
			break
		}
	}

	if len(course.Slides) == 0 {
		c.HTML(http.StatusOK, "employee/course_view.html", gin.H{
			"title":      course.Title,
			"userName":   c.MustGet("session_user_name"),
			"course":     course,
			"assignment": a,
		})
		return
	}
	co.assnSvc.UpdateCourseSlide(a.ID, slideIDParam)

	var prevID, nextID uint
	if currentIdx > 0 {
		prevID = course.Slides[currentIdx-1].ID
	}
	if currentIdx < len(course.Slides)-1 {
		nextID = course.Slides[currentIdx+1].ID
	}
	isLast := currentIdx == len(course.Slides)-1

	c.HTML(http.StatusOK, "employee/course_view.html", gin.H{
		"title":      course.Title,
		"userName":   c.MustGet("session_user_name"),
		"course":     course,
		"currentIdx": currentIdx,
		"slide":      course.Slides[currentIdx],
		"prevID":     prevID,
		"nextID":     nextID,
		"isLast":     isLast,
		"assignment": a,
	})
}

func (co *CourseController) CompleteCourse(c *gin.Context) {
	courseID := parseID(c, "id")
	userID := sessionUserID(c)
	co.assnSvc.CompleteCourse(userID, courseID)
	c.Redirect(http.StatusFound, "/employee/dashboard")
}

func parseID(c *gin.Context, param string) uint {
	var val string
	if param == "slide" {
		val = c.Query("slide")
	} else {
		val = c.Param(param)
	}
	id, _ := strconv.ParseUint(val, 10, 64)
	return uint(id)
}
