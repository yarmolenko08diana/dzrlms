package services

import (
	"lms/repositories"
	"gorm.io/gorm"
)

type Repos struct {
	Users       repositories.UserRepository
	Courses     repositories.CourseRepository
	Tests       repositories.TestRepository
	Assignments repositories.AssignmentRepository
}

type Container struct {
	Repos       Repos
	Courses     *CourseService
	Tests       *TestService
	Assignments *AssignmentService
	Notif       *NotificationService
	Upload      *UploadService
}

func NewContainer(gormDB *gorm.DB) *Container {
	userRepo   := repositories.NewUserRepository(gormDB)
	courseRepo := repositories.NewCourseRepository(gormDB)
	testRepo   := repositories.NewTestRepository(gormDB)
	assnRepo   := repositories.NewAssignmentRepository(gormDB)
	notifRepo  := repositories.NewNotificationRepository(gormDB)

	notif  := NewNotificationService(notifRepo)
	upload := NewUploadService("static/uploads", "/uploads")

	return &Container{
		Repos: Repos{
			Users:       userRepo,
			Courses:     courseRepo,
			Tests:       testRepo,
			Assignments: assnRepo,
		},
		Courses:     NewCourseService(courseRepo),
		Tests:       NewTestService(testRepo),
		Assignments: NewAssignmentService(assnRepo, userRepo, courseRepo, testRepo, notif),
		Notif:       notif,
		Upload:      upload,
	}
}
