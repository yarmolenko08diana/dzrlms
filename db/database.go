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

	seedSampleCourseAndTest()
}

func seedSampleCourseAndTest() {

	var course models.Course

	var courseCount int64
	DB.Model(&models.Course{}).Count(&courseCount)

	if courseCount == 0 {
		course = models.Course{
			Title:       "Введение в работу компании",
			Description: "Базовый курс для новых сотрудников: знакомство с компанией, правилами внутреннего распорядка и техникой безопасности.",
			Duration:    "30 минут",
			Published:   true,
		}
		DB.Create(&course)

		slides := []struct {
			Title   string
			Content string
		}{
			{
				Title: "Добро пожаловать в компанию",
				Content: "Здравствуйте и добро пожаловать в нашу компанию!\n\n" +
					"Этот курс поможет вам быстрее адаптироваться на новом месте: вы узнаете об основных правилах, структуре компании и технике безопасности.\n\n" +
					"Пройдите все главы курса, а затем — итоговый тест для проверки знаний.",
			},
			{
				Title: "Правила внутреннего трудового распорядка",
				Content: "Рабочий день начинается в 09:00 и заканчивается в 18:00, обед — с 13:00 до 14:00.\n\n" +
					"Об опоздании или болезни необходимо сообщать своему руководителю заранее.\n\n" +
					"Испытательный срок для новых сотрудников составляет 3 месяца.",
			},
			{
				Title: "Техника безопасности на рабочем месте",
				Content: "При обнаружении пожара немедленно сообщите о возгорании и следуйте плану эвакуации к ближайшему выходу.\n\n" +
					"При работе за компьютером делайте перерывы каждые 45–60 минут для отдыха глаз и разминки.\n\n" +
					"По всем вопросам охраны труда обращайтесь к специалисту по технике безопасности.",
			},
		}

		for i, sl := range slides {
			slide := models.Slide{CourseID: course.ID, OrderIndex: i, Title: sl.Title}
			DB.Create(&slide)

			DB.Create(&models.Block{
				SlideID:  slide.ID,
				Type:     "text",
				Content:  sl.Content,
				X:        60,
				Y:        60,
				W:        640,
				H:        320,
				ZIndex:   1,
				FontSize: 16,
			})
		}

		log.Println("Seeded sample course: Введение в работу компании")
	} else {
		DB.Where("title = ?", "Введение в работу компании").First(&course)
	}

	var testCount int64
	DB.Model(&models.Test{}).Count(&testCount)

	if testCount == 0 {
		var courseID *uint
		if course.ID != 0 {
			id := course.ID
			courseID = &id
		}

		test := models.Test{
			Title:       "Итоговый тест: Введение в работу компании",
			Description: "Проверьте свои знания по материалам вводного курса.",
			CourseID:    courseID,
			Published:   true,
		}
		DB.Create(&test)

		questions := []struct {
			Content string
			Answers []string
			Correct int
		}{
			{
				Content: "Во сколько начинается рабочий день в компании?",
				Answers: []string{"08:00", "09:00", "10:00", "Свободный график"},
				Correct: 1,
			},
			{
				Content: "Сколько длится испытательный срок для новых сотрудников?",
				Answers: []string{"1 месяц", "2 месяца", "3 месяца", "6 месяцев"},
				Correct: 2,
			},
			{
				Content: "Что нужно сделать при обнаружении пожара?",
				Answers: []string{
					"Продолжить работу, ничего не сообщая",
					"Сообщить о возгорании и следовать плану эвакуации",
					"Самостоятельно потушить пожар любыми средствами",
					"Подождать, пока пожар потухнет сам",
				},
				Correct: 1,
			},
			{
				Content: "Как часто рекомендуется делать перерывы при работе за компьютером?",
				Answers: []string{"Каждые 45–60 минут", "Один раз в день", "Раз в неделю", "Перерывы не нужны"},
				Correct: 0,
			},
			{
				Content: "Куда нужно обращаться по вопросам охраны труда?",
				Answers: []string{
					"К специалисту по технике безопасности",
					"В бухгалтерию",
					"В отдел маркетинга",
					"Никуда, решать самостоятельно",
				},
				Correct: 0,
			},
		}

		for qi, q := range questions {
			question := models.Question{TestID: test.ID, Content: q.Content, OrderIndex: qi}
			DB.Create(&question)

			for ai, ans := range q.Answers {
				DB.Create(&models.Answer{
					QuestionID: question.ID,
					Content:    ans,
					IsCorrect:  ai == q.Correct,
					OrderIndex: ai,
				})
			}
		}

		log.Println("Seeded sample test: Итоговый тест: Введение в работу компании")
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}