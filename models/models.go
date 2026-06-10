package models

import "time"

const (
	RoleAdmin    = "admin"
	RoleEmployee = "employee"

	StatusNotStarted = "not_started"
	StatusInProgress = "in_progress"
	StatusCompleted  = "completed"

	// Assignment target types
	AssignmentTypeCourse = "course"
	AssignmentTypeTest   = "test"
)

type User struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"not null"                 json:"name"`
	Email     string    `gorm:"unique;not null"          json:"email"`
	Password  string    `gorm:"not null"                 json:"-"`
	Role      string    `gorm:"not null;default:'employee'" json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Course struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Title       string    `gorm:"not null"                 json:"title"`
	Description string    `gorm:"type:text"                json:"description"`
	Duration    string    `json:"duration"`
	Published   bool      `gorm:"not null;default:false"   json:"published"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	Slides []Slide `gorm:"foreignKey:CourseID;constraint:OnDelete:CASCADE" json:"slides,omitempty"`
}

type Slide struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	CourseID   uint      `gorm:"not null;index"           json:"course_id"`
	OrderIndex int       `gorm:"not null;default:0"       json:"order_index"`
	Title      string    `json:"title"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	Blocks []Block `gorm:"foreignKey:SlideID;constraint:OnDelete:CASCADE" json:"blocks,omitempty"`
}

type Block struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	SlideID   uint      `gorm:"not null;index"           json:"slide_id"`
	Type      string    `gorm:"not null"                 json:"type"`
	Content   string    `gorm:"type:text"                json:"content"`
	X         float64   `gorm:"not null;default:0"       json:"x"`
	Y         float64   `gorm:"not null;default:0"       json:"y"`
	W         float64   `gorm:"not null;default:300"     json:"w"`
	H         float64   `gorm:"not null;default:120"     json:"h"`
	ZIndex    int       `gorm:"not null;default:0"       json:"z_index"`
	FontSize  int       `gorm:"not null;default:0"       json:"font_size"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Test struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	CourseID    *uint     `gorm:"index"                    json:"course_id"` // optional link to course
	Title       string    `gorm:"not null"                 json:"title"`
	Description string    `gorm:"type:text"                json:"description"`
	Published   bool      `gorm:"not null;default:false"   json:"published"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	Course    *Course    `gorm:"foreignKey:CourseID"                          json:"course,omitempty"`
	Questions []Question `gorm:"foreignKey:TestID;constraint:OnDelete:CASCADE" json:"questions,omitempty"`
}

type Question struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	TestID     uint      `gorm:"not null;index"           json:"test_id"`
	Content    string    `gorm:"type:text;not null"       json:"content"`
	MediaURL   string    `json:"media_url"` // optional image/video
	OrderIndex int       `gorm:"not null;default:0"       json:"order_index"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	Answers []Answer `gorm:"foreignKey:QuestionID;constraint:OnDelete:CASCADE" json:"answers,omitempty"`
}

type Answer struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	QuestionID uint      `gorm:"not null;index"           json:"question_id"`
	Content    string    `gorm:"type:text;not null"       json:"content"`
	MediaURL   string    `json:"media_url"`
	IsCorrect  bool      `gorm:"not null;default:false"   json:"is_correct"`
	OrderIndex int       `gorm:"not null;default:0"       json:"order_index"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type Assignment struct {
	ID         uint      `gorm:"primaryKey;autoIncrement"      json:"id"`
	UserID     uint      `gorm:"not null;index"                json:"user_id"`
	TargetType string    `gorm:"not null;default:'course'"     json:"target_type"` // "course" | "test"
	TargetID   uint      `gorm:"not null;default:0;index"      json:"target_id"`
	Status     string    `gorm:"not null;default:'not_started'" json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
}

type CourseProgress struct {
	ID              uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	AssignmentID    uint      `gorm:"not null;uniqueIndex"     json:"assignment_id"`
	CurrentSlideID  *uint     `json:"current_slide_id"`
	CompletedSlides int       `gorm:"not null;default:0"       json:"completed_slides"`
	UpdatedAt       time.Time `json:"updated_at"`

	Assignment   Assignment `gorm:"foreignKey:AssignmentID"  json:"assignment,omitempty"`
	CurrentSlide *Slide     `gorm:"foreignKey:CurrentSlideID;constraint:OnDelete:CASCADE" json:"current_slide,omitempty"`
}

type TestProgress struct {
	ID                   uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	AssignmentID         uint       `gorm:"not null;uniqueIndex"     json:"assignment_id"`
	CurrentQuestionIndex int        `gorm:"not null;default:0"       json:"current_question_index"`
	Score                int        `gorm:"not null;default:0"       json:"score"`
	StartedAt            *time.Time `json:"started_at"`
	FinishedAt           *time.Time `json:"finished_at"`
	UpdatedAt            time.Time  `json:"updated_at"`

	Assignment       Assignment        `gorm:"foreignKey:AssignmentID;constraint:OnDelete:CASCADE" json:"assignment,omitempty"`
	IncorrectAnswers []IncorrectAnswer `gorm:"foreignKey:TestProgressID;constraint:OnDelete:CASCADE" json:"incorrect_answers,omitempty"`
}

type IncorrectAnswer struct {
	ID             uint `gorm:"primaryKey;autoIncrement" json:"id"`
	TestProgressID uint `gorm:"not null;index"           json:"test_progress_id"`
	QuestionID     uint `gorm:"not null;index"           json:"question_id"`
	AnswerID       uint `gorm:"not null;index"           json:"answer_id"`

	Question Question `gorm:"foreignKey:QuestionID" json:"question,omitempty"`
	Answer   Answer   `gorm:"foreignKey:AnswerID"   json:"answer,omitempty"`
}

type Notification struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint      `gorm:"not null;index"           json:"user_id"`
	Message   string    `gorm:"type:text;not null"       json:"message"`
	Read      bool      `gorm:"not null;default:false"   json:"read"`
	CreatedAt time.Time `json:"created_at"`

	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
}

type DashboardStats struct {
	TotalEmployees          int64
	TotalCourses            int64
	TotalTests              int64
	ActiveCourseAssignments int64
	ActiveTestAssignments   int64
	CompletedAssignments    int64
}
