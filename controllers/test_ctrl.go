package controllers

import (
	"encoding/json"
	"net/http"

	"lms/models"
	"lms/repositories"
	"lms/services"
	"strconv"

	"github.com/gin-gonic/gin"
)

type TestController struct {
	svc     *services.TestService
	assnSvc *services.AssignmentService
	users   repositories.UserRepository
	upload  *services.UploadService
}

func NewTestController(
	svc *services.TestService,
	assnSvc *services.AssignmentService,
	users repositories.UserRepository,
	upload *services.UploadService,
) *TestController {
	return &TestController{svc: svc, assnSvc: assnSvc, users: users, upload: upload}
}

func (tc *TestController) List(c *gin.Context) {
	tests, _ := tc.svc.List()
	c.HTML(http.StatusOK, "admin/tests.html", gin.H{
		"title":    "Тесты",
		"tests":    tests,
		"active":   "tests",
		"userName": c.MustGet("session_user_name"),
	})
}

func (tc *TestController) NewEditor(c *gin.Context) {
	employees, _ := tc.users.FindEmployees()
	c.HTML(http.StatusOK, "admin/test_editor.html", gin.H{
		"title":     "Создать тест",
		"active":    "tests",
		"userName":  c.MustGet("session_user_name"),
		"employees": employees,
	})
}

func (tc *TestController) EditEditor(c *gin.Context) {
	id := parseID(c, "id")
	test, err := tc.svc.Get(id)
	if err != nil {
		c.Redirect(http.StatusFound, "/admin/tests")
		return
	}
	employees, _ := tc.users.FindEmployees()
	testJSON, _ := json.Marshal(test)
	c.HTML(http.StatusOK, "admin/test_editor.html", gin.H{
		"title":     "Редактировать тест",
		"active":    "tests",
		"userName":  c.MustGet("session_user_name"),
		"test":      test,
		"testJSON":  string(testJSON),
		"employees": employees,
		"editing":   true,
	})
}

func (tc *TestController) Save(c *gin.Context) {
	testID := parseID(c, "id")
	var payload services.SaveTestPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	test, err := tc.svc.Upsert(testID, payload)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": test.ID})
}

func (tc *TestController) Publish(c *gin.Context) {
	testID := parseID(c, "id")
	var req struct {
		AssignAll bool   `json:"assign_all"`
		UserIDs   []uint `json:"user_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := tc.svc.Publish(testID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := tc.assnSvc.Assign(services.AssignRequest{
		TargetType: models.AssignmentTypeTest,
		TargetID:   testID,
		AssignAll:  req.AssignAll,
		UserIDs:    req.UserIDs,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (tc *TestController) Delete(c *gin.Context) {
	tc.svc.Delete(parseID(c, "id"))
	c.Redirect(http.StatusFound, "/admin/tests")
}

func (tc *TestController) UploadMedia(c *gin.Context) {
	fh, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no file"})
		return
	}
	url, err := tc.upload.Save(fh)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"url": url})
}

func (tc *TestController) EmployeeView(c *gin.Context) {
	testID := parseID(c, "id")
	userID := sessionUserID(c)

	a, prog, err := tc.assnSvc.OpenTest(userID, testID)
	if err != nil {
		c.Redirect(http.StatusFound, "/employee/dashboard")
		return
	}
	test, err := tc.svc.Get(testID)
	if err != nil {
		c.Redirect(http.StatusFound, "/employee/dashboard")
		return
	}

	if len(test.Questions) == 0 {
		c.Redirect(http.StatusFound, "/employee/dashboard")
		return
	}

	var idx int
	if q := c.Query("idx"); q != "" {
		idx, _ = strconv.Atoi(q)
	} else {
		idx = prog.CurrentQuestionIndex
	}
	if idx < 0 {
		idx = 0
	}
	if idx >= len(test.Questions) {
		idx = len(test.Questions) - 1
	}

	c.HTML(http.StatusOK, "employee/test_view.html", gin.H{
		"title":      test.Title,
		"userName":   c.MustGet("session_user_name"),
		"test":       test,
		"assignment": a,
		"progress":   prog,
		"currentIdx": idx,
		"question":   test.Questions[idx],
		"totalCount": len(test.Questions),
		"isLast":     idx == len(test.Questions)-1,
	})
}

func (tc *TestController) SubmitAnswer(c *gin.Context) {
	testID := parseID(c, "id")
	userID := sessionUserID(c)

	qID, _ := strconv.ParseUint(c.PostForm("question_id"), 10, 64)
	aID, _ := strconv.ParseUint(c.PostForm("answer_id"), 10, 64)
	nextIdx, _ := strconv.Atoi(c.PostForm("next_idx"))
	isCorrect := c.PostForm("is_correct") == "true"

	if err := tc.assnSvc.SubmitAnswer(
		userID,
		testID,
		uint(qID),
		uint(aID),
		isCorrect,
		nextIdx,
	); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if c.PostForm("is_last") == "true" {
		tc.assnSvc.CompleteTest(userID, testID)
		c.Redirect(http.StatusFound, "/employee/tests/"+idStr(testID)+"/results")
		return
	}

	c.Redirect(http.StatusFound, "/employee/tests/"+idStr(testID))
}

func (tc *TestController) CompleteTest(c *gin.Context) {
	testID := parseID(c, "id")
	userID := sessionUserID(c)
	tc.assnSvc.CompleteTest(userID, testID)
	c.Redirect(http.StatusFound, "/employee/tests/"+idStr(testID)+"/results")
}

func (tc *TestController) Results(c *gin.Context) {
	testID := parseID(c, "id")
	userID := sessionUserID(c)

	a, err := tc.assnSvc.FindByUserAndTarget(userID, testID)
	if err != nil {
		c.Redirect(http.StatusFound, "/employee/dashboard")
		return
	}

	test, err := tc.svc.Get(testID)
	if err != nil {
		c.Redirect(http.StatusFound, "/employee/dashboard")
		return
	}

	_, prog, answers, err := tc.assnSvc.GetTestResults(a.ID)
	if err != nil {
		c.Redirect(http.StatusFound, "/employee/dashboard")
		return
	}

	answersByQ := make(map[uint]models.TestAnswer, len(answers))
	for _, ta := range answers {
		answersByQ[ta.QuestionID] = ta
	}

	c.HTML(http.StatusOK, "employee/test_results.html", gin.H{
		"title":      "Результаты теста",
		"userName":   c.MustGet("session_user_name"),
		"test":       test,
		"assignment": a,
		"progress":   prog,
		"answers":    answers,
		"answersByQ": answersByQ,
	})
}

func idStr(id uint) string {
	return strconv.FormatUint(uint64(id), 10)
}