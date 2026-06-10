package services

import (
	"fmt"
	"lms/models"
	"lms/repositories"
)

type AssignmentService struct {
	assignments repositories.AssignmentRepository
	users       repositories.UserRepository
	courses     repositories.CourseRepository
	tests       repositories.TestRepository
	notif       *NotificationService
}

func NewAssignmentService(
	a repositories.AssignmentRepository,
	u repositories.UserRepository,
	c repositories.CourseRepository,
	t repositories.TestRepository,
	n *NotificationService,
) *AssignmentService {
	return &AssignmentService{
		assignments: a,
		users:       u,
		courses:     c,
		tests:       t,
		notif:       n,
	}
}

type AssignRequest struct {
	TargetType string // "course" | "test"
	TargetID   uint
	UserIDs    []uint
	AssignAll  bool
}

func (s *AssignmentService) Assign(req AssignRequest) error {

	if req.TargetType != models.AssignmentTypeCourse &&
		req.TargetType != models.AssignmentTypeTest {
		return fmt.Errorf("invalid target_type: %s", req.TargetType)
	}

	targetTitle, err := s.resolveTitle(req.TargetType, req.TargetID)
	if err != nil {
		return err
	}

	userIDs := req.UserIDs
	if req.AssignAll {
		employees, err := s.users.FindEmployees()
		if err != nil {
			return err
		}
		for _, u := range employees {
			userIDs = append(userIDs, u.ID)
		}
	}

	for _, uid := range userIDs {

		existing, _ := s.assignments.FindByUserAndTarget(uid, req.TargetType, req.TargetID)
		if existing != nil {
			continue
		}

		a := &models.Assignment{
			UserID:     uid,
			TargetType: req.TargetType,
			TargetID:   req.TargetID,
			Status:     models.StatusNotStarted,
		}

		if err := s.assignments.Create(a); err != nil {
			return err
		}

		user, err := s.users.FindByID(uid)
		if err != nil {
			continue
		}

		switch req.TargetType {
		case models.AssignmentTypeCourse:
			s.notif.NotifyCourseAssigned(*user, targetTitle)
		case models.AssignmentTypeTest:
			s.notif.NotifyTestAssigned(*user, targetTitle)
		}
	}

	return nil
}

type AssignBatchRequest struct {
	TargetTypes []string
	CourseID    uint
	TestID      uint
	UserIDs     []uint
	AssignAll   bool
}

func (s *AssignmentService) AssignBatch(req AssignBatchRequest) error {

	for _, tt := range req.TargetTypes {

		var targetID uint

		switch tt {
		case models.AssignmentTypeCourse:
			targetID = req.CourseID

		case models.AssignmentTypeTest:
			targetID = req.TestID

		default:
			continue
		}

		if targetID == 0 {
			continue
		}

		err := s.Assign(AssignRequest{
			TargetType: tt,
			TargetID:   targetID,
			UserIDs:    req.UserIDs,
			AssignAll:  req.AssignAll,
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func (s *AssignmentService) resolveTitle(targetType string, targetID uint) (string, error) {
	switch targetType {
	case models.AssignmentTypeCourse:
		c, err := s.courses.FindByID(targetID)
		if err != nil {
			return "", err
		}
		return c.Title, nil

	case models.AssignmentTypeTest:
		t, err := s.tests.FindByID(targetID)
		if err != nil {
			return "", err
		}
		return t.Title, nil
	}
	return "", nil
}

func (s *AssignmentService) List(targetType, status string) ([]models.Assignment, error) {
	return s.assignments.FindAll(targetType, status)
}

func (s *AssignmentService) Delete(id uint) error {
	return s.assignments.Delete(id)
}

func (s *AssignmentService) OpenCourse(userID, courseID uint) (*models.Assignment, *models.CourseProgress, error) {
	a, err := s.assignments.FindByUserAndTarget(userID, models.AssignmentTypeCourse, courseID)
	if err != nil {
		return nil, nil, fmt.Errorf("не назначен")
	}

	prog, err := s.assignments.FindOrCreateCourseProgress(a.ID)
	if err != nil {
		return nil, nil, err
	}

	if a.Status == models.StatusNotStarted {
		s.setAssignmentStatus(a.ID, models.StatusInProgress)
	}

	return a, prog, nil
}

func (s *AssignmentService) UpdateCourseSlide(assignmentID uint, slideID uint) error {
	prog, err := s.assignments.FindOrCreateCourseProgress(assignmentID)
	if err != nil {
		return err
	}
	prog.CurrentSlideID = &slideID
	return s.assignments.UpdateCourseProgress(prog)
}

func (s *AssignmentService) CompleteCourse(userID, courseID uint) error {
	a, err := s.assignments.FindByUserAndTarget(userID, models.AssignmentTypeCourse, courseID)
	if err != nil {
		return fmt.Errorf("не назначен")
	}
	if a.Status != models.StatusInProgress {
		return fmt.Errorf("неверный статус")
	}
	return s.setAssignmentStatus(a.ID, models.StatusCompleted)
}

func (s *AssignmentService) OpenTest(userID, testID uint) (*models.Assignment, *models.TestProgress, error) {
	a, err := s.assignments.FindByUserAndTarget(userID, models.AssignmentTypeTest, testID)
	if err != nil {
		return nil, nil, fmt.Errorf("не назначен")
	}

	prog, err := s.assignments.FindOrCreateTestProgress(a.ID)
	if err != nil {
		return nil, nil, err
	}

	if a.Status == models.StatusNotStarted {
		s.setAssignmentStatus(a.ID, models.StatusInProgress)
	}

	return a, prog, nil
}

func (s *AssignmentService) SubmitAnswer(
	userID, testID uint,
	questionID, answerID uint,
	isCorrect bool,
	nextQuestionIndex int,
) error {

	a, err := s.assignments.FindByUserAndTarget(userID, models.AssignmentTypeTest, testID)
	if err != nil {
		return fmt.Errorf("не назначен")
	}

	prog, err := s.assignments.FindOrCreateTestProgress(a.ID)
	if err != nil {
		return err
	}

	prog.CurrentQuestionIndex = nextQuestionIndex

	if isCorrect {
		prog.Score++
	} else {
		s.assignments.AddIncorrectAnswer(&models.IncorrectAnswer{
			TestProgressID: prog.ID,
			QuestionID:     questionID,
			AnswerID:       answerID,
		})
	}

	return s.assignments.UpdateTestProgress(prog)
}

func (s *AssignmentService) CompleteTest(userID, testID uint) error {
	a, err := s.assignments.FindByUserAndTarget(userID, models.AssignmentTypeTest, testID)
	if err != nil {
		return fmt.Errorf("не назначен")
	}
	return s.setAssignmentStatus(a.ID, models.StatusCompleted)
}

func (s *AssignmentService) setAssignmentStatus(id uint, status string) error {
	a, err := s.assignments.FindByID(id)
	if err != nil {
		return err
	}
	a.Status = status
	return s.assignments.Update(a)
}
