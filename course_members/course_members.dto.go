package course_members

import "errors"

type GetUserByCourseMemberIDDTO struct {
	ID int64
}

func (dto GetUserByCourseMemberIDDTO) Validate() error {
	if dto.ID == 0 {
		return errors.New("id is required")
	}
	return nil
}

type GetCourseWithMembersDTO struct {
	ID int64
}

func (dto GetCourseWithMembersDTO) Validate() error {
	if dto.ID == 0 {
		return errors.New("id is required")
	}
	return nil
}

type JoinCourseDTO struct {
	CourseID int64
	UserID   int64
}

func (dto JoinCourseDTO) Validate() error {
	if dto.CourseID == 0 {
		return errors.New("course id is required")
	}
	if dto.UserID == 0 {
		return errors.New("user id is required")
	}
	return nil
}
