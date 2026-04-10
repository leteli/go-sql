package courses

import "errors"

type CreateCourseDTO struct {
	Slug  string
	Title string
	Price int
}

func (dto CreateCourseDTO) Validate() error {
	if dto.Slug == "" {
		return errors.New("slug is required")
	}
	if dto.Title == "" {
		return errors.New("title is required")
	}
	if dto.Price < 0 {
		return errors.New("price must be >= 0")
	}
	return nil
}

type ListCoursesDTO struct {
	OrderKey string
	Limit    int
	Offset   int
}

func (dto ListCoursesDTO) Validate() error {
	if dto.Limit < 0 {
		return errors.New("limit must be >= 0")
	}
	if dto.Offset < 0 {
		return errors.New("offset must be >= 0")
	}
	return nil
}

type FindCoursesByIDsDTO struct {
	IDs []int
}

func (dto FindCoursesByIDsDTO) Validate() error {
	if len(dto.IDs) == 0 {
		return errors.New("ids are required")
	}
	return nil
}
