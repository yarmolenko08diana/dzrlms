package repositories

import "lms/models"

type UserRepository interface {
	FindAll() ([]models.User, error)
	FindByID(id uint) (*models.User, error)
	FindByEmail(email string) (*models.User, error)
	FindEmployees() ([]models.User, error)
	Create(u *models.User) error
	Update(u *models.User) error
	Delete(id uint) error
}

type CourseRepository interface {
	FindAll() ([]models.Course, error)
	FindByID(id uint) (*models.Course, error)
	FindWithSlides(id uint) (*models.Course, error)
	Create(c *models.Course) error
	Update(c *models.Course) error
	Delete(id uint) error
	// Slide operations
	UpsertSlide(s *models.Slide) error
	DeleteSlidesNotIn(courseID uint, keepIDs []uint) error
	// Block operations
	UpsertBlock(b *models.Block) error
	DeleteBlocksNotIn(slideID uint, keepIDs []uint) error
}

type TestRepository interface {
	FindAll() ([]models.Test, error)
	FindByID(id uint) (*models.Test, error)
	FindWithQuestions(id uint) (*models.Test, error)
	Create(t *models.Test) error
	Update(t *models.Test) error
	Delete(id uint) error
	// Question operations
	UpsertQuestion(q *models.Question) error
	DeleteQuestionsNotIn(testID uint, keepIDs []uint) error
	// Answer operations
	UpsertAnswer(a *models.Answer) error
	DeleteAnswersNotIn(questionID uint, keepIDs []uint) error
}

type AssignmentRepository interface {
	FindAll(targetType, status string) ([]models.Assignment, error)
	FindByUser(userID uint) ([]models.Assignment, error)
	FindByID(id uint) (*models.Assignment, error)
	FindByUserAndTarget(userID uint, targetType string, targetID uint) (*models.Assignment, error)
	Create(a *models.Assignment) error
	Update(a *models.Assignment) error
	Delete(id uint) error
	// Progress
	FindOrCreateCourseProgress(assignmentID uint) (*models.CourseProgress, error)
	UpdateCourseProgress(p *models.CourseProgress) error
	FindOrCreateTestProgress(assignmentID uint) (*models.TestProgress, error)
	UpdateTestProgress(p *models.TestProgress) error
	AddIncorrectAnswer(ia *models.IncorrectAnswer) error
	UpsertTestAnswer(ta *models.TestAnswer) error
	FindTestAnswers(testProgressID uint) ([]models.TestAnswer, error)
}

type NotificationRepository interface {
	Create(n *models.Notification) error
	FindForUser(userID uint) ([]models.Notification, error)
	MarkRead(id uint) error
	MarkAllReadForUser(userID uint) error
}