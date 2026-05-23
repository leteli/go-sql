package orders

import "errors"

type ListOrdersDTO struct {
	Limit  int64
	Offset int64
}

func (dto ListOrdersDTO) Validate() error {
	if dto.Limit < 0 {
		return errors.New("limit must be >= 0")
	}
	if dto.Offset < 0 {
		return errors.New("offset must be >= 0")
	}
	return nil
}

type GetUserByOrderIDDTO struct {
	ID int64
}

func (dto GetUserByOrderIDDTO) Validate() error {
	if dto.ID == 0 {
		return errors.New("id is required")
	}
	return nil
}
