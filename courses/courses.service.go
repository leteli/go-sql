package courses

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"go-sql/internal/storage"
	"strings"
)

var ErrNotFound = errors.New("course not found")

func CreateCourse(ctx context.Context, db *sql.DB, dto CreateCourseDTO) (int, error) {
	if err := dto.Validate(); err != nil {
		return 0, err
	}
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
	if err := dto.Validate(); err != nil {
		return nil, err
	}
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
	defer rows.Close()

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
	if err := dto.Validate(); err != nil {
		return nil, err
	}
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
	defer rows.Close()

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

func UpdateCoursePrices(ctx context.Context, db *sql.DB, dto UpdateCoursePricesDTO) ([]storage.Course, error) {
	if err := dto.Validate(); err != nil {
		return nil, err
	}
	stmt, err := db.PrepareContext(ctx, `
		UPDATE courses
		SET price = ?
		WHERE id = ?
		RETURNING id, slug, title, price
	`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	var courses []storage.Course

	for _, p := range dto.Prices {
		var c storage.Course
		err := stmt.QueryRowContext(ctx, p.Price, p.ID).Scan(&c.ID, &c.Slug, &c.Title, &c.Price)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, fmt.Errorf("%w; course id=%d", ErrNotFound, p.ID)
			}
			return nil, err
		}
		courses = append(courses, c)
	}
	return courses, nil
}

func ListCoursesByMaxPrices(ctx context.Context, db *sql.DB, dto ListCoursesByMaxPricesDTO) (map[int][]storage.Course, error) {
	if err := dto.Validate(); err != nil {
		return nil, err
	}
	stmt, err := db.PrepareContext(ctx, `
		SELECT id, slug, title, price
		FROM courses
		WHERE price <= ?
	`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	var coursesBelowLimit = make(map[int][]storage.Course, len(dto.Prices))

	for _, p := range dto.Prices {
		rows, err := stmt.QueryContext(ctx, p)
		if err != nil {
			return nil, err
		}
		var courses []storage.Course
		for rows.Next() {
			var c storage.Course
			if err := rows.Scan(&c.ID, &c.Slug, &c.Title, &c.Price); err != nil {
				rows.Close()
				return nil, err
			}
			courses = append(courses, c)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return nil, err
		}
		coursesBelowLimit[p] = courses
	}
	return coursesBelowLimit, nil
}

func BulkWriteCourses(ctx context.Context, db *sql.DB, dto BulkWriteCoursesDTO) error {
	if err := dto.Validate(); err != nil {
		return err
	}
	stmt, err := db.PrepareContext(ctx, `
		INSERT INTO courses (slug, title, price)
		VALUES (?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, c := range dto.Courses {
		_, err := stmt.ExecContext(ctx, c.Slug, c.Title, c.Price)
		if err != nil {
			return err
		}
	}
	return nil
}
