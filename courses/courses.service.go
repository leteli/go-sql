package courses

import (
	"context"
	"database/sql"
	"fmt"
	"go-sql/internal/storage"
	"strings"
)

func CreateCourse(ctx context.Context, db *sql.DB, dto CreateCourseDTO) (int, error) {
	query := `
	INSERT INTO courses
	(slug, title, price)
	VALUES (?, ?, ?)
	`
	res, err := db.ExecContext(ctx, query, dto.Slug, dto.Title, dto.Price)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

var allowedOrders = map[string]string{
	"id_asc":     "id ASC",
	"title_asc":  "title ASC",
	"price_asc":  "price ASC",
	"price_desc": "price DESC",
}

func ListCourses(ctx context.Context, db *sql.DB, dto ListCoursesDTO) ([]storage.Course, error) {
	order, ok := allowedOrders[dto.OrderKey]
	if !ok {
		order = allowedOrders["id_asc"]
	}
	query := fmt.Sprintf(
		`SELECT id, slug, title, price FROM courses ORDER BY %s LIMIT %d OFFSET %d`,
		order,
		dto.Limit,
		dto.Offset,
	)
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	var courses []storage.Course
	for rows.Next() {
		var c storage.Course
		if err := rows.Scan(&c.ID, &c.Slug, &c.Title, &c.Price); err != nil {
			return nil, err
		}
		courses = append(courses, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return courses, nil
}

func FindCoursesByIDs(ctx context.Context, db *sql.DB, dto FindCoursesByIDsDTO) ([]storage.Course, error) {
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(dto.IDs)), ",")

	query := fmt.Sprintf(`
	SELECT id, slug, title, price
	FROM courses
	WHERE id IN (%s)`, placeholders)

	args := make([]any, len(dto.IDs))
	for i, v := range dto.IDs {
		args[i] = v
	}

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	var courses []storage.Course
	for rows.Next() {
		var c storage.Course
		if err := rows.Scan(&c.ID, &c.Slug, &c.Title, &c.Price); err != nil {
			return nil, err
		}
		courses = append(courses, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return courses, nil
}
