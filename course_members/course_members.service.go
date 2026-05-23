package course_members

import (
	"context"
	"database/sql"
	"errors"
	"go-sql/courses"
	coursemembersdb "go-sql/internal/db/course_members"
	coursesdb "go-sql/internal/db/courses"
	ordersdb "go-sql/internal/db/orders"
	usersdb "go-sql/internal/db/users"
	"go-sql/users"
	"go-sql/utils"
)

func GetUserByCourseMemberID(ctx context.Context, q coursemembersdb.Querier, dto GetUserByCourseMemberIDDTO) (coursemembersdb.GetUserByCourseMemberIDRow, error) {
	if err := dto.Validate(); err != nil {
		return coursemembersdb.GetUserByCourseMemberIDRow{}, err
	}
	return q.GetUserByCourseMemberID(ctx, dto.ID)
} // TODO: handle sql.Null* types

func GetCourseWithMembers(ctx context.Context, q coursemembersdb.Querier, dto GetCourseWithMembersDTO) (coursemembersdb.GetCourseWithMembersRow, error) {
	if err := dto.Validate(); err != nil {
		return coursemembersdb.GetCourseWithMembersRow{}, err
	}
	return q.GetCourseWithMembers(ctx, dto.ID)
} // TODO: handle sql.Null* types

func JoinCourse(
	ctx context.Context,
	conn *sql.DB,
	dto JoinCourseDTO,
) error {
	if err := dto.Validate(); err != nil {
		return err
	}

	return utils.WithTx(ctx, conn, func(tx *sql.Tx) error {
		cmQ := coursemembersdb.New(tx)
		oQ := ordersdb.New(tx)
		cQ := coursesdb.New(tx)
		uQ := usersdb.New(tx)
		c, err := cQ.FindCourseByID(ctx, dto.CourseID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return courses.ErrNotFound
			}
			return err
		}
		if _, err := uQ.FindUserByID(ctx, dto.UserID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return users.ErrNotFound
			}
			return err
		}
		err = cmQ.CreateCourseMember(ctx, coursemembersdb.CreateCourseMemberParams{
			UserID: sql.NullInt64{
				Int64: dto.UserID,
				Valid: true,
			},
			CourseID: sql.NullInt64{
				Int64: dto.CourseID,
				Valid: true,
			},
		})
		if err != nil {
			return err
		}

		return oQ.CreateOrder(ctx, ordersdb.CreateOrderParams{
			UserID: sql.NullInt64{
				Int64: dto.UserID,
				Valid: true,
			},
			CourseID: sql.NullInt64{
				Int64: dto.CourseID,
				Valid: true,
			},
			AmountCents: c.Price,
		})
	})
}
