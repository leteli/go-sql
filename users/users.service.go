package users

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	usersdb "go-sql/internal/db/users"
	"go-sql/internal/storage"
	"strings"
)

var allowedOrders = map[string]string{
	"id_asc":     "id ASC",
	"title_asc":  "title ASC",
	"price_asc":  "price ASC",
	"price_desc": "price DESC",
}

var ErrNotFound = errors.New("user not found")

func CreateUser(ctx context.Context, q usersdb.Querier, dto CreateUserDTO) (int64, error) {
	if err := dto.Validate(); err != nil {
		return 0, err
	}
	params := usersdb.CreateUserParams{
		Email: dto.Email,
	}
	if dto.Name != nil {
		params.Name = sql.NullString{String: *dto.Name, Valid: true}
	}
	if dto.Age != nil {
		params.Age = sql.NullInt64{Int64: *dto.Age, Valid: true}
	}
	return q.CreateUser(ctx, params)
}

func CreateUserRaw(ctx context.Context, db *sql.DB, dto CreateUserDTO) (int64, error) {
	if err := dto.Validate(); err != nil {
		return 0, err
	}
	query := `
		INSERT INTO users
		(email, name, age)
		VALUES (?, ?, ?)
	`
	res, err := db.ExecContext(ctx, query, dto.Email, dto.Name, dto.Age)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func UpdateUser(ctx context.Context, q usersdb.Querier, dto UpdateUserDTO) (storage.User, error) {
	if err := dto.Validate(); err != nil {
		return storage.User{}, err
	}
	params := usersdb.UpdateUserParams{
		ID: dto.ID,
	}
	if dto.Email != nil {
		params.Email = sql.NullString{
			String: *dto.Email,
			Valid:  true,
		}
	}
	if dto.Name != nil {
		params.Name = sql.NullString{
			String: *dto.Name,
			Valid:  true,
		}
	}
	if dto.Age != nil {
		params.Age = sql.NullInt64{
			Int64: *dto.Age,
			Valid: true,
		}
	}
	u, err := q.UpdateUser(ctx, params)
	if err != nil {
		return storage.User{}, err
	}
	return toUserResponse(u), nil
}

func UpdateUserRaw(ctx context.Context, db *sql.DB, dto UpdateUserDTO) (storage.User, error) {
	if err := dto.Validate(); err != nil {
		return storage.User{}, err
	}
	updateValues := []string{}
	args := []any{}

	if dto.Email != nil {
		updateValues = append(updateValues, "email = ?")
		args = append(args, *dto.Email)
	}
	if dto.Name != nil {
		updateValues = append(updateValues, "name = ?")
		args = append(args, *dto.Name)
	}
	if dto.Age != nil {
		updateValues = append(updateValues, "age = ?")
		args = append(args, *dto.Age)
	}

	query := fmt.Sprintf(`
		UPDATE users
		SET %s
		WHERE users.id = ?
		RETURNING id, email, name, age, created_at
	`, strings.Join(updateValues, ", "))

	args = append(args, dto.ID)

	var u storage.User

	err := db.QueryRowContext(ctx, query, args...).Scan(&u.ID, &u.Email, &u.Name, &u.Age, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return u, ErrNotFound
		}
		return u, err
	}
	return u, nil
}

func FindUserByID(ctx context.Context, q usersdb.Querier, dto FindUserByIDDTO) (storage.User, error) {
	if err := dto.Validate(); err != nil {
		return storage.User{}, err
	}
	u, err := q.FindUserByID(ctx, dto.ID)
	if err != nil {
		return storage.User{}, err
	}
	return toUserResponse(u), nil
}

func FindUserByIDRaw(ctx context.Context, db *sql.DB, dto FindUserByIDDTO) (storage.User, error) {
	if err := dto.Validate(); err != nil {
		return storage.User{}, err
	}
	query := `
		SELECT id, email, name, age, created_at
		FROM users
		WHERE id = ?
	`
	var u storage.User

	err := db.QueryRowContext(ctx, query, dto.ID).Scan(&u.ID, &u.Email, &u.Name, &u.Age, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return storage.User{}, ErrNotFound
		}
		return storage.User{}, err
	}
	return u, nil
}

func ListUsers(ctx context.Context, q usersdb.Querier, dto ListUsersDTO) ([]storage.User, error) {
	if err := dto.Validate(); err != nil {
		return nil, err
	}
	params := usersdb.ListUsersParams{
		Limit:  dto.Limit,
		Offset: dto.Offset,
	}
	dbUsers, err := q.ListUsers(ctx, params)
	if err != nil {
		return nil, err
	}
	users := make([]storage.User, len(dbUsers))
	for i, u := range dbUsers {
		users[i] = toUserResponse(u)
	}
	return users, nil
}

func ListUsersRaw(ctx context.Context, db *sql.DB, dto ListUsersDTO) ([]storage.User, error) {
	if err := dto.Validate(); err != nil {
		return nil, err
	}
	order, ok := allowedOrders[dto.OrderKey]
	if !ok {
		order = allowedOrders["id_asc"]
	}
	query := fmt.Sprintf(
		`
		SELECT id, email, name, age, created_at
		FROM users
		ORDER BY %s
		LIMIT %d
		OFFSET %d
		`,
		order,
		dto.Limit,
		dto.Offset,
	)
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []storage.User
	for rows.Next() {
		var u storage.User
		if err := rows.Scan(&u.ID, &u.Email, &u.Name, &u.Age, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

func DeleteUser(ctx context.Context, q usersdb.Querier, dto DeleteUserDTO) error {
	if err := dto.Validate(); err != nil {
		return err
	}
	return q.DeleteUser(ctx, dto.ID)
}

func DeleteUserRaw(ctx context.Context, db *sql.DB, dto DeleteUserDTO) (int64, error) {
	if err := dto.Validate(); err != nil {
		return 0, err
	}
	query := `
		DELETE FROM users
		WHERE id = ?
	`
	res, err := db.ExecContext(ctx, query, dto.ID)
	if err != nil {
		return 0, err
	}
	rCount, err := res.RowsAffected()
	if rCount == 0 {
		return 0, sql.ErrNoRows
	}
	return rCount, nil
}

func toUserResponse(u usersdb.User) storage.User {
	r := storage.User{
		ID:        u.ID,
		Email:     u.Email,
		CreatedAt: u.CreatedAt.Time,
	}
	if u.Name.Valid {
		r.Name = &u.Name.String
	}
	if u.Age.Valid {
		r.Age = &u.Age.Int64
	}
	return r
}
