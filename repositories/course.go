package repositories

import (
	"gorm.io/gorm"
	"lms/models"
)

type courseRepo struct{ db *gorm.DB }

func NewCourseRepository(db *gorm.DB) CourseRepository {
	return &courseRepo{db: db}
}

func (r *courseRepo) FindAll() ([]models.Course, error) {
	var courses []models.Course
	return courses, r.db.Find(&courses).Error
}

func (r *courseRepo) FindByID(id uint) (*models.Course, error) {
	var c models.Course
	return &c, r.db.First(&c, id).Error
}

func (r *courseRepo) FindWithSlides(id uint) (*models.Course, error) {
	var c models.Course
	err := r.db.First(&c, id).Error
	if err != nil {
		return nil, err
	}
	var slides []models.Slide
	r.db.Where("course_id = ?", id).Order("order_index").Find(&slides)
	for i := range slides {
		var blocks []models.Block
		r.db.Where("slide_id = ?", slides[i].ID).Order("z_index").Find(&blocks)
		slides[i].Blocks = blocks
	}
	c.Slides = slides
	return &c, nil
}

func (r *courseRepo) Create(c *models.Course) error { return r.db.Create(c).Error }
func (r *courseRepo) Update(c *models.Course) error { return r.db.Save(c).Error }

func (r *courseRepo) Delete(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var slideIDs []uint
		if err := tx.Model(&models.Slide{}).
			Where("course_id = ?", id).
			Pluck("id", &slideIDs).Error; err != nil {
			return err
		}

		if len(slideIDs) > 0 {
			if err := tx.Model(&models.CourseProgress{}).
				Where("current_slide_id IN ?", slideIDs).
				Update("current_slide_id", gorm.Expr("NULL")).Error; err != nil {
				return err
			}
		}

		var assnIDs []uint
		if err := tx.Model(&models.Assignment{}).
			Where("target_type = ? AND target_id = ?", "course", id).
			Pluck("id", &assnIDs).Error; err != nil {
			return err
		}

		if len(assnIDs) > 0 {
			if err := tx.Where("assignment_id IN ?", assnIDs).
				Delete(&models.CourseProgress{}).Error; err != nil {
				return err
			}
			if err := tx.Where("id IN ?", assnIDs).
				Delete(&models.Assignment{}).Error; err != nil {
				return err
			}
		}

		if len(slideIDs) > 0 {
			if err := tx.Where("slide_id IN ?", slideIDs).
				Delete(&models.Block{}).Error; err != nil {
				return err
			}
			if err := tx.Where("course_id = ?", id).
				Delete(&models.Slide{}).Error; err != nil {
				return err
			}
		}
		return tx.Delete(&models.Course{}, id).Error
	})
}

func (r *courseRepo) UpsertSlide(s *models.Slide) error {
	if s.ID == 0 {
		return r.db.Create(s).Error
	}
	return r.db.Save(s).Error
}

func (r *courseRepo) DeleteSlidesNotIn(courseID uint, keepIDs []uint) error {
	q := r.db.Where("course_id = ?", courseID)
	if len(keepIDs) > 0 {
		q = q.Where("id NOT IN ?", keepIDs)
	}
	return q.Delete(&models.Slide{}).Error
}

func (r *courseRepo) UpsertBlock(b *models.Block) error {
	if b.ID == 0 {
		return r.db.Create(b).Error
	}
	return r.db.Save(b).Error
}

func (r *courseRepo) DeleteBlocksNotIn(slideID uint, keepIDs []uint) error {
	q := r.db.Where("slide_id = ?", slideID)
	if len(keepIDs) > 0 {
		q = q.Where("id NOT IN ?", keepIDs)
	}
	return q.Delete(&models.Block{}).Error
}
