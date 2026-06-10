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

func (r *userRepo) Create(u *models.User) error  { return r.db.Create(u).Error }
func (r *userRepo) Update(u *models.User) error  { return r.db.Save(u).Error }
func (r *userRepo) Delete(id uint) error         { return r.db.Delete(&models.User{}, id).Error }
