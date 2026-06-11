package repositories

import (
	"lms/models"
	"gorm.io/gorm"
)

type notifRepo struct{ db *gorm.DB }

func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notifRepo{db: db}
}

func (r *notifRepo) Create(n *models.Notification) error { return r.db.Create(n).Error }

func (r *notifRepo) FindForUser(userID uint) ([]models.Notification, error) {
	var notifs []models.Notification
	return notifs, r.db.Where("user_id = ?", userID).
		Order("created_at desc").Limit(50).Find(&notifs).Error
}

func (r *notifRepo) MarkRead(id uint) error {
	return r.db.Model(&models.Notification{}).Where("id = ?", id).Update("read", true).Error
}

func (r *notifRepo) MarkAllReadForUser(userID uint) error {
	return r.db.Model(&models.Notification{}).Where("user_id = ? AND read = ?", userID, false).Update("read", true).Error
}
