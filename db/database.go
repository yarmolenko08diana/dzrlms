package db

import (
	"log"
	"os"

	"lms/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Connect() {

	dsn := "host=" + getEnv("DB_HOST", "localhost") +
		" user=" + getEnv("DB_USER", "postgres") +
		" password=" + getEnv("DB_PASSWORD", "postgres") +
		" dbname=" + getEnv("DB_NAME", "lms_db") +
		" port=" + getEnv("DB_PORT", "5432") +
		" sslmode=disable TimeZone=UTC client_encoding=UTF8"

	var err error

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("Failed to get sql.DB: %v", err)
	}

	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("Database ping failed: %v", err)
	}

	log.Println("Database connected.")
}

func Migrate() {

	err := DB.AutoMigrate(
		&models.User{},
		&models.Course{},
		&models.Slide{},
		&models.Block{},
		&models.Test{},
		&models.Question{},
		&models.Answer{},

		&models.Assignment{},

		&models.CourseProgress{},
		&models.TestProgress{},
		&models.IncorrectAnswer{},
		&models.TestAnswer{},

		&models.Notification{},
	)

	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("Database migrated.")
}

func Seed() {

	var adminCount int64
	DB.Model(&models.User{}).Where("role = ?", models.RoleAdmin).Count(&adminCount)

	if adminCount == 0 {
		hash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)

		DB.Create(&models.User{
			Name:     "System Admin",
			Email:    "admin@company.com",
			Password: string(hash),
			Role:     models.RoleAdmin,
		})

		log.Println("Seeded admin: admin@company.com / admin123")
	}

	var empCount int64
	DB.Model(&models.User{}).Where("role = ?", models.RoleEmployee).Count(&empCount)

	if empCount == 0 {

		names := []string{"Диана Ярмоленко"}
		emails := []string{"yarmolenko08diana@gmail.com"}

		for i := range names {
			hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

			DB.Create(&models.User{
				Name:     names[i],
				Email:    emails[i],
				Password: string(hash),
				Role:     models.RoleEmployee,
			})
		}

		log.Println("Seeded employees")
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}