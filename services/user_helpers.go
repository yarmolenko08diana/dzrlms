package services

import "lms/models"

func NewUserModel(name, email, hashedPassword string) models.User {
	return models.User{
		Name:     name,
		Email:    email,
		Password: hashedPassword,
		Role:     models.RoleEmployee,
	}
}
