package controllers

import (
	"net/http"

	"lms/services"

	"github.com/gin-gonic/gin"
)

type AssignmentController struct {
	svc *services.AssignmentService
}

func NewAssignmentController(svc *services.AssignmentService) *AssignmentController {
	return &AssignmentController{svc: svc}
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

	c.HTML(http.StatusOK, "admin/assignments.html", gin.H{
		"title":        "Назначения",
		"rows":         list,
		"active":       "assignments",
		"userName":     c.MustGet("session_user_name"),
		"filterType":   targetType,
		"filterStatus": status,
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