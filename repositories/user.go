package repositories

import (
	"lms/models"
	"gorm.io/gorm"
)

type userRepo struct{ db *gorm.DB }

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) FindAll() ([]models.User, error) {
	var users []models.User
	return users, r.db.Find(&users).Error
}

func (r *userRepo) FindByID(id uint) (*models.User, error) {
	var u models.User
	return &u, r.db.First(&u, id).Error
}

func (r *userRepo) FindByEmail(email string) (*models.User, error) {
	var u models.User
	return &u, r.db.Where("email = ?", email).First(&u).Error
}

func (r *userRepo) FindEmployees() ([]models.User, error) {
	var users []models.User
	return users, r.db.Where("role = ?", models.RoleEmployee).Find(&users).Error
}

func (r *userRepo) Create(u *models.User) error { return r.db.Create(u).Error }
func (r *userRepo) Update(u *models.User) error { return r.db.Save(u).Error }

func (r *userRepo) Delete(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var assnIDs []uint
		if err := tx.Model(&models.Assignment{}).
			Where("user_id = ?", id).
			Pluck("id", &assnIDs).Error; err != nil {
			return err
		}

		if len(assnIDs) > 0 {
			var progressIDs []uint
			if err := tx.Model(&models.TestProgress{}).
				Where("assignment_id IN ?", assnIDs).
				Pluck("id", &progressIDs).Error; err != nil {
				return err
			}

			if len(progressIDs) > 0 {
				if err := tx.Where("test_progress_id IN ?", progressIDs).
					Delete(&models.IncorrectAnswer{}).Error; err != nil {
					return err
				}
				if err := tx.Where("test_progress_id IN ?", progressIDs).
					Delete(&models.TestAnswer{}).Error; err != nil {
					return err
				}
			}

			if err := tx.Where("assignment_id IN ?", assnIDs).
				Delete(&models.CourseProgress{}).Error; err != nil {
				return err
			}
			if err := tx.Where("assignment_id IN ?", assnIDs).
				Delete(&models.TestProgress{}).Error; err != nil {
				return err
			}
			if err := tx.Where("id IN ?", assnIDs).
				Delete(&models.Assignment{}).Error; err != nil {
				return err
			}
		}

		if err := tx.Where("user_id = ?", id).
			Delete(&models.Notification{}).Error; err != nil {
			return err
		}

		return tx.Delete(&models.User{}, id).Error
	})
}
