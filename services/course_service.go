package services

import (
	"fmt"
	"lms/models"
	"lms/repositories"
)

type CourseService struct {
	courses repositories.CourseRepository
	upload  *UploadService
}

func NewCourseService(courses repositories.CourseRepository, upload *UploadService) *CourseService {
	return &CourseService{courses: courses, upload: upload}
}

func (s *CourseService) List() ([]models.Course, error) {
	return s.courses.FindAll()
}

func (s *CourseService) Get(id uint) (*models.Course, error) {
	return s.courses.FindWithSlides(id)
}

type SaveCoursePayload struct {
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Duration    string         `json:"duration"`
	Slides      []SlidePayload `json:"slides"`
}

type SlidePayload struct {
	ID         uint           `json:"id"`
	OrderIndex int            `json:"order_index"`
	Title      string         `json:"title"`
	Blocks     []BlockPayload `json:"blocks"`
}

type BlockPayload struct {
	ID       uint    `json:"id"`
	Type     string  `json:"type"`
	Content  string  `json:"content"`
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	W        float64 `json:"w"`
	H        float64 `json:"h"`
	ZIndex   int     `json:"z_index"`
	FontSize int     `json:"font_size"`
}

func (s *CourseService) Upsert(courseID uint, p SaveCoursePayload) (*models.Course, error) {
	if p.Title == "" {
		return nil, fmt.Errorf("title is required")
	}

	var course models.Course
	if courseID == 0 {
		course = models.Course{Title: p.Title, Description: p.Description, Duration: p.Duration}
		if err := s.courses.Create(&course); err != nil {
			return nil, err
		}
	} else {
		c, err := s.courses.FindByID(courseID)
		if err != nil {
			return nil, err
		}
		c.Title = p.Title
		c.Description = p.Description
		c.Duration = p.Duration
		if err := s.courses.Update(c); err != nil {
			return nil, err
		}
		course = *c
	}

	var keepSlideIDs []uint
	for i, sp := range p.Slides {
		slide := models.Slide{
			ID:         sp.ID,
			CourseID:   course.ID,
			OrderIndex: i,
			Title:      sp.Title,
		}
		if err := s.courses.UpsertSlide(&slide); err != nil {
			return nil, err
		}
		keepSlideIDs = append(keepSlideIDs, slide.ID)

		var keepBlockIDs []uint
		for _, bp := range sp.Blocks {
			block := models.Block{
				ID:       bp.ID,
				SlideID:  slide.ID,
				Type:     bp.Type,
				Content:  bp.Content,
				X: bp.X, Y: bp.Y, W: bp.W, H: bp.H,
				ZIndex:   bp.ZIndex,
				FontSize: bp.FontSize,
			}
			if err := s.courses.UpsertBlock(&block); err != nil {
				return nil, err
			}
			keepBlockIDs = append(keepBlockIDs, block.ID)
		}
		if err := s.courses.DeleteBlocksNotIn(slide.ID, keepBlockIDs); err != nil {
			return nil, err
		}
	}
	if err := s.courses.DeleteSlidesNotIn(course.ID, keepSlideIDs); err != nil {
		return nil, err
	}

	return &course, nil
}

func (s *CourseService) Publish(courseID uint) error {
	c, err := s.courses.FindByID(courseID)
	if err != nil {
		return err
	}
	c.Published = true
	return s.courses.Update(c)
}

func (s *CourseService) Delete(id uint) error {
	if c, err := s.courses.FindWithSlides(id); err == nil {
		for _, slide := range c.Slides {
			for _, b := range slide.Blocks {
				if b.Type == "image" || b.Type == "video" {
					s.upload.DeleteFile(b.Content)
				}
			}
		}
	}
	return s.courses.Delete(id)
}