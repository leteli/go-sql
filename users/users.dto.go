package users

import "errors"

type CreateUserDTO struct {
	Email string
	Name  *string
	Age   *int64
}

func (dto CreateUserDTO) Validate() error {
	if dto.Email == "" {
		return errors.New("email is required")
	}
	return nil
}

type UpdateUserDTO struct {
	ID    int64
	Email *string
	Name  *string
	Age   *int64
}

func (dto UpdateUserDTO) Validate() error {
	if dto.ID == 0 {
		return errors.New("id is required")
	}
	if dto.Email == nil && dto.Name == nil && dto.Age == nil {
		return errors.New("no update parameters defined")
	}
	return nil
}

type FindUserByIDDTO struct {
	ID int64
}

func (dto FindUserByIDDTO) Validate() error {
	if dto.ID == 0 {
		return errors.New("id is required")
	}
	return nil
}

type ListUsersDTO struct {
	OrderKey string
	Limit    int64
	Offset   int64
}

func (dto ListUsersDTO) Validate() error {
	if dto.Limit < 0 {
		return errors.New("limit must be >= 0")
	}
	if dto.Offset < 0 {
		return errors.New("offset must be >= 0")
	}
	return nil
}

type DeleteUserDTO struct {
	ID int64
}

func (dto DeleteUserDTO) Validate() error {
	if dto.ID == 0 {
		return errors.New("no id provided")
	}
	return nil
}
