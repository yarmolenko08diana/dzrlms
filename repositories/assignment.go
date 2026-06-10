package repositories

import (
	"time"
	"lms/models"
	"gorm.io/gorm"
)

type assignmentRepo struct{ db *gorm.DB }

func NewAssignmentRepository(db *gorm.DB) AssignmentRepository {
	return &assignmentRepo{db: db}
}

func (r *assignmentRepo) FindByUser(userID uint) ([]models.Assignment, error) {
	var list []models.Assignment
	return list, r.db.Where("user_id = ?", userID).Find(&list).Error
}

func (r *assignmentRepo) FindAll(targetType, status string) ([]models.Assignment, error) {
	var list []models.Assignment
	q := r.db.Preload("User")
	if targetType != "" {
		q = q.Where("target_type = ?", targetType)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	return list, q.Find(&list).Error
}

func (r *assignmentRepo) FindByID(id uint) (*models.Assignment, error) {
	var a models.Assignment
	return &a, r.db.Preload("User").First(&a, id).Error
}

func (r *assignmentRepo) FindByUserAndTarget(userID uint, targetType string, targetID uint) (*models.Assignment, error) {
	var a models.Assignment
	err := r.db.Where("user_id = ? AND target_type = ? AND target_id = ?", userID, targetType, targetID).
		First(&a).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *assignmentRepo) Create(a *models.Assignment) error { return r.db.Create(a).Error }
func (r *assignmentRepo) Update(a *models.Assignment) error  { return r.db.Save(a).Error }

func (r *assignmentRepo) Delete(id uint) error {
	r.db.Where("assignment_id = ?", id).Delete(&models.CourseProgress{})
	r.db.Where("assignment_id = ?", id).Delete(&models.TestProgress{})
	return r.db.Delete(&models.Assignment{}, id).Error
}

func (r *assignmentRepo) FindOrCreateCourseProgress(assignmentID uint) (*models.CourseProgress, error) {
	var p models.CourseProgress
	err := r.db.Where("assignment_id = ?", assignmentID).First(&p).Error
	if err != nil {
		p = models.CourseProgress{AssignmentID: assignmentID}
		err = r.db.Create(&p).Error
	}
	return &p, err
}

func (r *assignmentRepo) UpdateCourseProgress(p *models.CourseProgress) error {
	p.UpdatedAt = time.Now()
	return r.db.Save(p).Error
}

func (r *assignmentRepo) FindOrCreateTestProgress(assignmentID uint) (*models.TestProgress, error) {
	var p models.TestProgress
	err := r.db.Where("assignment_id = ?", assignmentID).First(&p).Error
	if err != nil {
		p = models.TestProgress{AssignmentID: assignmentID}
		err = r.db.Create(&p).Error
	}
	return &p, err
}

func (r *assignmentRepo) UpdateTestProgress(p *models.TestProgress) error {
	p.UpdatedAt = time.Now()
	return r.db.Save(p).Error
}

func (r *assignmentRepo) AddIncorrectAnswer(ia *models.IncorrectAnswer) error {
	return r.db.Create(ia).Error
}
