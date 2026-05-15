package users

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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

func CreateUser(ctx context.Context, db *sql.DB, dto CreateUserDTO) (int, error) {
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
	return int(id), nil
}

func UpdateUser(ctx context.Context, db *sql.DB, dto UpdateUserDTO) (storage.User, error) {
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

func FindUserByID(ctx context.Context, db *sql.DB, dto FindUserByIDDTO) (storage.User, error) {
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

func ListUsers(ctx context.Context, db *sql.DB, dto ListUsersDTO) ([]storage.User, error) {
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

func DeleteUser(ctx context.Context, db *sql.DB, dto DeleteUserDTO) (int64, error) {
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
