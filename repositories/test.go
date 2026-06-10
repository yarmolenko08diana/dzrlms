package repositories

import (
	"lms/models"
	"gorm.io/gorm"
)

type testRepo struct{ db *gorm.DB }

func NewTestRepository(db *gorm.DB) TestRepository {
	return &testRepo{db: db}
}

func (r *testRepo) FindAll() ([]models.Test, error) {
	var tests []models.Test
	return tests, r.db.Find(&tests).Error
}

func (r *testRepo) FindByID(id uint) (*models.Test, error) {
	var t models.Test
	return &t, r.db.First(&t, id).Error
}

func (r *testRepo) FindWithQuestions(id uint) (*models.Test, error) {
	var t models.Test
	err := r.db.First(&t, id).Error
	if err != nil {
		return nil, err
	}
	var questions []models.Question
	r.db.Where("test_id = ?", id).Order("order_index").Find(&questions)
	for i := range questions {
		var answers []models.Answer
		r.db.Where("question_id = ?", questions[i].ID).Order("order_index").Find(&answers)
		questions[i].Answers = answers
	}
	t.Questions = questions
	return &t, nil
}

func (r *testRepo) Create(t *models.Test) error { return r.db.Create(t).Error }
func (r *testRepo) Update(t *models.Test) error { return r.db.Save(t).Error }
func (r *testRepo) Delete(id uint) error {
	return r.db.Select("Questions.Answers").Delete(&models.Test{}, id).Error
}

func (r *testRepo) UpsertQuestion(q *models.Question) error {
	if q.ID == 0 {
		return r.db.Create(q).Error
	}
	return r.db.Save(q).Error
}

func (r *testRepo) DeleteQuestionsNotIn(testID uint, keepIDs []uint) error {
	q := r.db.Where("test_id = ?", testID)
	if len(keepIDs) > 0 {
		q = q.Where("id NOT IN ?", keepIDs)
	}
	return q.Delete(&models.Question{}).Error
}

func (r *testRepo) UpsertAnswer(a *models.Answer) error {
	if a.ID == 0 {
		return r.db.Create(a).Error
	}
	return r.db.Save(a).Error
}

func (r *testRepo) DeleteAnswersNotIn(questionID uint, keepIDs []uint) error {
	q := r.db.Where("question_id = ?", questionID)
	if len(keepIDs) > 0 {
		q = q.Where("id NOT IN ?", keepIDs)
	}
	return q.Delete(&models.Answer{}).Error
}
