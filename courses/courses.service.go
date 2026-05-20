package courses

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	coursesdb "go-sql/internal/db/courses"
	"go-sql/internal/storage"
	"go-sql/utils"
	"strings"
)

var ErrNotFound = errors.New("course not found")

func CreateCourse(ctx context.Context, q coursesdb.Querier, dto CreateCourseDTO) (int64, error) {
	if err := dto.Validate(); err != nil {
		return 0, err
	}
	params := coursesdb.CreateCourseParams{
		Slug:  dto.Slug,
		Title: dto.Title,
		Price: dto.Price,
	}
	return q.CreateCourse(ctx, params)
}

func CreateCourseRaw(ctx context.Context, db *sql.DB, dto CreateCourseDTO) (int, error) {
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

func ListCourses(ctx context.Context, q coursesdb.Querier, dto ListCoursesDTO) ([]coursesdb.Course, error) {
	if err := dto.Validate(); err != nil {
		return nil, err
	}
	params := coursesdb.ListCoursesParams{
		// OrderBy: dto.OrderKey, // NB: use bare sql for dynamic ORDER BY
		Limit:  dto.Limit,
		Offset: dto.Offset,
	}
	return q.ListCourses(ctx, params)
}

var allowedOrdersRaw = map[string]string{
	"id_asc":     "id ASC",
	"title_asc":  "title ASC",
	"price_asc":  "price ASC",
	"price_desc": "price DESC",
}

func ListCoursesRaw(ctx context.Context, db *sql.DB, dto ListCoursesDTO) ([]storage.Course, error) {
	if err := dto.Validate(); err != nil {
		return nil, err
	}
	order, ok := allowedOrdersRaw[dto.OrderKey]
	if !ok {
		order = allowedOrdersRaw["id_asc"]
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

func FindCoursesByIDs(ctx context.Context, q coursesdb.Querier, dto FindCoursesByIDsDTO) ([]coursesdb.Course, error) {
	if err := dto.Validate(); err != nil {
		return nil, err
	}
	return q.FindCoursesByIDs(ctx, dto.IDs)
}

func FindCoursesByIDsRaw(ctx context.Context, db *sql.DB, dto FindCoursesByIDsDTO) ([]storage.Course, error) {
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

func UpdateCoursePrices(ctx context.Context, db *sql.DB, dto UpdateCoursePricesDTO) ([]coursesdb.Course, error) {
	if err := dto.Validate(); err != nil {
		return nil, err
	}
	courses := make([]coursesdb.Course, 0, len(dto.Prices))
	update := func(tx *sql.Tx) error {
		coursesPQ, err := coursesdb.Prepare(ctx, tx)
		if err != nil {
			return err
		}
		defer coursesPQ.Close()

		for _, p := range dto.Prices {

			course, err := coursesPQ.UpdateCoursePrice(
				ctx,
				coursesdb.UpdateCoursePriceParams{
					ID:    p.ID,
					Price: p.Price,
				})
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return fmt.Errorf("%w; course id=%d", ErrNotFound, p.ID)
				}
				return err
			}
			courses = append(courses, course)
		}
		return nil
	}
	if err := utils.WithTx(ctx, db, update); err != nil {
		return nil, fmt.Errorf("transaction error %w", err)
	}
	return courses, nil
}

func UpdateCoursePricesRaw(ctx context.Context, db *sql.DB, dto UpdateCoursePricesDTO) ([]storage.Course, error) {
	if err := dto.Validate(); err != nil {
		return nil, err
	}
	courses := make([]storage.Course, 0, len(dto.Prices))

	update := func(tx *sql.Tx) error {
		stmt, err := tx.PrepareContext(ctx, `
		UPDATE courses
		SET price = ?
		WHERE id = ?
		RETURNING id, slug, title, price
	`)
		if err != nil {
			return err
		}
		defer stmt.Close()

		for _, p := range dto.Prices {
			var c storage.Course
			err := stmt.QueryRowContext(ctx, p.Price, p.ID).Scan(&c.ID, &c.Slug, &c.Title, &c.Price)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return fmt.Errorf("%w; course id=%d", ErrNotFound, p.ID)
				}
				return err
			}
			courses = append(courses, c)
		}
		return nil
	}
	if err := utils.WithTx(ctx, db, update); err != nil {
		return nil, fmt.Errorf("transaction error %w", err)
	}
	return courses, nil
}

func ListCoursesByMaxPrices(ctx context.Context, q coursesdb.Querier, dto ListCoursesByMaxPricesDTO) (map[int64][]coursesdb.Course, error) {
	if err := dto.Validate(); err != nil {
		return nil, err
	}
	var coursesBelowLimit = make(map[int64][]coursesdb.Course, len(dto.Prices))
	for _, p := range dto.Prices {
		courses, err := q.ListCoursesByMaxPrice(ctx, p)
		if err != nil {
			return nil, err
		}
		coursesBelowLimit[p] = courses
	}
	return coursesBelowLimit, nil
}

func ListCoursesByMaxPricesRaw(ctx context.Context, db *sql.DB, dto ListCoursesByMaxPricesDTO) (map[int64][]storage.Course, error) {
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
	var coursesBelowLimit = make(map[int64][]storage.Course, len(dto.Prices))

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

func BulkWriteCourses(ctx context.Context, db *sql.DB, dto BulkWriteCoursesDTO) (OpStatus, error) {
	if err := dto.Validate(); err != nil {
		return OpStatus{Status: "failed"}, err
	}
	bulkWrite := func(tx *sql.Tx) error {
		coursesPQ, err := coursesdb.Prepare(ctx, tx)
		if err != nil {
			return err
		}
		defer coursesPQ.Close()

		for _, c := range dto.Courses {
			params := coursesdb.CreateCourseParams{
				Slug:  c.Slug,
				Title: c.Title,
				Price: c.Price,
			}
			if _, err := coursesPQ.CreateCourse(ctx, params); err != nil {
				return err
			}
		}

		return nil
	}
	err := utils.WithTx(ctx, db, bulkWrite)
	if err != nil {
		return OpStatus{Status: "failed"}, fmt.Errorf("transaction error %w", err)
	}
	return OpStatus{Status: "success"}, nil
}

func BulkWriteCoursesRaw(ctx context.Context, db *sql.DB, dto BulkWriteCoursesDTO) (OpStatus, error) {
	if err := dto.Validate(); err != nil {
		return OpStatus{Status: "failed"}, err
	}
	bulkWrite := func(tx *sql.Tx) error {
		stmt, err := tx.PrepareContext(ctx, `
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
	err := utils.WithTx(ctx, db, bulkWrite)
	if err != nil {
		return OpStatus{Status: "failed"}, fmt.Errorf("transaction error %w", err)
	}
	return OpStatus{Status: "success"}, nil
}
