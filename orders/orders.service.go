package orders

import (
	"context"
	"database/sql"
	"errors"
	ordersdb "go-sql/internal/db/orders"
)

var ErrNotFound = errors.New("order not found")

func ListOrders(ctx context.Context, q ordersdb.Querier, dto ListOrdersDTO) ([]ordersdb.ListOrdersRow, error) {
	if err := dto.Validate(); err != nil {
		return nil, err
	}
	return q.ListOrders(ctx, ordersdb.ListOrdersParams{
		Limit:  dto.Limit,
		Offset: dto.Offset,
	}) // TODO: handle sql.Null* types
}

func GetUserByOrderID(ctx context.Context, q ordersdb.Querier, dto GetUserByOrderIDDTO) (ordersdb.GetUserByOrderIDRow, error) {
	if err := dto.Validate(); err != nil {
		return ordersdb.GetUserByOrderIDRow{}, err
	}
	u, err := q.GetUserByOrderID(ctx, dto.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ordersdb.GetUserByOrderIDRow{}, ErrNotFound
		}
		return ordersdb.GetUserByOrderIDRow{}, err
	}
	return u, nil
} // TODO: handle sql.Null* types
