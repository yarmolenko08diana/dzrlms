package services

import (
	"fmt"
	"lms/models"
	"lms/repositories"
)

type TestService struct {
	tests repositories.TestRepository
}

func NewTestService(tests repositories.TestRepository) *TestService {
	return &TestService{tests: tests}
}

func (s *TestService) List() ([]models.Test, error) {
	return s.tests.FindAll()
}

func (s *TestService) Get(id uint) (*models.Test, error) {
	return s.tests.FindWithQuestions(id)
}

type SaveTestPayload struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	CourseID    *uint             `json:"course_id"`
	Questions   []QuestionPayload `json:"questions"`
}

type QuestionPayload struct {
	ID       uint            `json:"id"`
	Content  string          `json:"content"`
	MediaURL string          `json:"media_url"`
	Answers  []AnswerPayload `json:"answers"`
}

type AnswerPayload struct {
	ID        uint   `json:"id"`
	Content   string `json:"content"`
	MediaURL  string `json:"media_url"`
	IsCorrect bool   `json:"is_correct"`
}

func (s *TestService) Upsert(testID uint, p SaveTestPayload) (*models.Test, error) {
	if p.Title == "" {
		return nil, fmt.Errorf("title is required")
	}

	var test models.Test
	if testID == 0 {
		test = models.Test{Title: p.Title, Description: p.Description, CourseID: p.CourseID}
		if err := s.tests.Create(&test); err != nil {
			return nil, err
		}
	} else {
		t, err := s.tests.FindByID(testID)
		if err != nil {
			return nil, err
		}
		t.Title = p.Title
		t.Description = p.Description
		t.CourseID = p.CourseID
		if err := s.tests.Update(t); err != nil {
			return nil, err
		}
		test = *t
	}

	var keepQIDs []uint
	for i, qp := range p.Questions {
		q := models.Question{
			ID:         qp.ID,
			TestID:     test.ID,
			Content:    qp.Content,
			MediaURL:   qp.MediaURL,
			OrderIndex: i,
		}
		if err := s.tests.UpsertQuestion(&q); err != nil {
			return nil, err
		}
		keepQIDs = append(keepQIDs, q.ID)

		var keepAIDs []uint
		for j, ap := range qp.Answers {
			a := models.Answer{
				ID:         ap.ID,
				QuestionID: q.ID,
				Content:    ap.Content,
				MediaURL:   ap.MediaURL,
				IsCorrect:  ap.IsCorrect,
				OrderIndex: j,
			}
			if err := s.tests.UpsertAnswer(&a); err != nil {
				return nil, err
			}
			keepAIDs = append(keepAIDs, a.ID)
		}
		if err := s.tests.DeleteAnswersNotIn(q.ID, keepAIDs); err != nil {
			return nil, err
		}
	}
	if err := s.tests.DeleteQuestionsNotIn(test.ID, keepQIDs); err != nil {
		return nil, err
	}

	return &test, nil
}

func (s *TestService) Publish(testID uint) error {
	t, err := s.tests.FindByID(testID)
	if err != nil {
		return err
	}
	t.Published = true
	return s.tests.Update(t)
}

func (s *TestService) Delete(id uint) error {
	return s.tests.Delete(id)
}
