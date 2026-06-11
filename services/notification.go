package services

import (
	"fmt"
	"log"
	"lms/models"
	"lms/repositories"
)

type NotificationService struct {
	repo repositories.NotificationRepository
}

func NewNotificationService(repo repositories.NotificationRepository) *NotificationService {
	return &NotificationService{repo: repo}
}

func (s *NotificationService) NotifyCourseAssigned(user models.User, courseTitle string) {
	msg := fmt.Sprintf("Вам назначен новый курс: %s", courseTitle)
	if err := s.repo.Create(&models.Notification{UserID: user.ID, Message: msg}); err != nil {
		log.Printf("[NOTIF ERROR] %v", err)
	}
	// Email stub — replace with real net/smtp implementation
	log.Printf("[EMAIL STUB] To: %s <%s> | %s", user.Name, user.Email, msg)
}

func (s *NotificationService) NotifyTestAssigned(user models.User, testTitle string) {
	msg := fmt.Sprintf("Вам назначен новый тест: %s", testTitle)
	if err := s.repo.Create(&models.Notification{UserID: user.ID, Message: msg}); err != nil {
		log.Printf("[NOTIF ERROR] %v", err)
	}
	log.Printf("[EMAIL STUB] To: %s <%s> | %s", user.Name, user.Email, msg)
}

func (s *NotificationService) ForUser(userID uint) ([]models.Notification, error) {
	return s.repo.FindForUser(userID)
}

func (s *NotificationService) MarkRead(id uint) error {
	return s.repo.MarkRead(id)
}

func (s *NotificationService) MarkAllRead(userID uint) error {
	return s.repo.MarkAllReadForUser(userID)
}