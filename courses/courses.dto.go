package courses

import "errors"

type CreateCourseDTO struct {
	Slug  string `json:"slug"`
	Title string `json:"title"`
	Price int    `json:"price"`
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

type CoursePrice struct {
	ID    int `json:"id"`
	Price int `json:"price"`
}

type UpdateCoursePricesDTO struct {
	Prices []CoursePrice
}

func (dto UpdateCoursePricesDTO) Validate() error {
	if len(dto.Prices) == 0 {
		return errors.New("course prices are required")
	}
	return nil
}

type ListCoursesByMaxPricesDTO struct {
	Prices []int
}

func (dto ListCoursesByMaxPricesDTO) Validate() error {
	if len(dto.Prices) == 0 {
		return errors.New("course prices are required")
	}
	return nil
}

type NewCourse = CreateCourseDTO

type BulkWriteCoursesDTO struct {
	Courses []NewCourse
}

func (dto BulkWriteCoursesDTO) Validate() error {
	if len(dto.Courses) == 0 {
		return errors.New("courses data is required")
	}
	return nil
}

type OpStatus struct {
	Status string `json:"status"`
}
