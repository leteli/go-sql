package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"go-sql/courses"
	"go-sql/users"
	"strconv"

	"encoding/json"

	coursesdb "go-sql/internal/db/courses"
	usersdb "go-sql/internal/db/users"

	"github.com/urfave/cli/v3"
	_ "modernc.org/sqlite"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	conn, err := sql.Open("sqlite", "file:data.db?_foreign_keys=on&_busy_timeout=5000")
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer conn.Close()

	if err := conn.PingContext(ctx); err != nil {
		log.Fatalf("ping db: %v", err)
	}
	usersQ := usersdb.New(conn)
	coursesQ := coursesdb.New(conn)
	coursesPQ, err := coursesdb.Prepare(ctx, conn)
	if err != nil {
		log.Fatalf("cannot use prepared statements: %v", err)
	}
	defer coursesPQ.Close()

	cmd := &cli.Command{
		Name:  "db-tool",
		Usage: "CLI for testing course DB operations",
		Commands: []*cli.Command{
			createCourseCommand(coursesQ),
			listCoursesCommand(coursesQ),
			findCoursesByIDsCommand(coursesQ),
			updateCoursePricesCommand(conn),
			listCoursesByMaxPricesCommand(coursesPQ),
			bulkWriteCoursesCommand(conn),
			createUserCommand(usersQ),
			updateUserCommand(usersQ),
			findUserByIDCommand(usersQ),
			listUsersCommand(usersQ),
			deleteUserCommand(usersQ),
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func createCourseCommand(q coursesdb.Querier) *cli.Command {
	return &cli.Command{
		Name:  "create-course",
		Usage: "create a new course",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "slug",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "title",
				Required: true,
			},
			&cli.Int64Flag{
				Name:     "price",
				Required: true,
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
			defer cancel()

			dto := courses.CreateCourseDTO{
				Slug:  c.String("slug"),
				Title: c.String("title"),
				Price: c.Int64("price"),
			}
			res, err := courses.CreateCourse(ctx, q, dto)
			if err != nil {
				return err
			}

			return printJSON(res)
		},
	}
}

func listCoursesCommand(q coursesdb.Querier) *cli.Command {
	return &cli.Command{
		Name:  "list-courses",
		Usage: "list courses",
		Flags: []cli.Flag{
			// &cli.StringFlag{
			// 	Name:  "order",
			// 	Value: "id_asc",
			// },
			&cli.Int64Flag{
				Name:  "limit",
				Value: 10,
			},
			&cli.Int64Flag{
				Name:  "offset",
				Value: 0,
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
			defer cancel()

			dto := courses.ListCoursesDTO{
				// OrderKey: c.String("order"),
				Limit:  c.Int64("limit"),
				Offset: c.Int64("offset"),
			}

			res, err := courses.ListCourses(ctx, q, dto)
			if err != nil {
				return err
			}

			return printJSON(res)
		},
	}
}

func findCoursesByIDsCommand(q coursesdb.Querier) *cli.Command {
	return &cli.Command{
		Name:  "find-courses-by-ids",
		Usage: "find courses by comma-separated ids",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "ids",
				Usage:    "comma-separated ids, e.g. 1,2,3",
				Required: true,
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
			defer cancel()

			ids, err := parseInt64Strings(c.String("ids"))
			if err != nil {
				return err
			}
			dto := courses.FindCoursesByIDsDTO{
				IDs: ids,
			}

			res, err := courses.FindCoursesByIDs(ctx, q, dto)
			if err != nil {
				return err
			}

			return printJSON(res)
		},
	}
}

func updateCoursePricesCommand(db *sql.DB) *cli.Command {
	return &cli.Command{
		Name:  "update-course-prices",
		Usage: "update course prices from JSON file",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "file",
				Usage:    "path to JSON file",
				Required: true,
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
			defer cancel()

			data, err := os.ReadFile(c.String("file"))
			if err != nil {
				return err
			}

			var prices []courses.CoursePrice
			if err := json.Unmarshal(data, &prices); err != nil {
				return err
			}

			dto := courses.UpdateCoursePricesDTO{
				Prices: prices,
			}

			res, err := courses.UpdateCoursePrices(ctx, db, dto)
			if err != nil {
				return err
			}

			return printJSON(res)
		},
	}
}

func listCoursesByMaxPricesCommand(q coursesdb.Querier) *cli.Command {
	return &cli.Command{
		Name:  "list-courses-by-max-prices",
		Usage: "list courses by max pricess",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "prices",
				Usage:    "comma-separated integer prices, e.g 100,1000,400",
				Required: true,
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
			defer cancel()

			prices, err := parseInt64Strings(c.String("prices"))
			if err != nil {
				return err
			}

			dto := courses.ListCoursesByMaxPricesDTO{
				Prices: prices,
			}

			res, err := courses.ListCoursesByMaxPrices(ctx, q, dto)
			if err != nil {
				return err
			}

			return printJSON(res)
		},
	}
}

func bulkWriteCoursesCommand(db *sql.DB) *cli.Command {
	return &cli.Command{
		Name:  "bulk-write-courses",
		Usage: "create multiple courses from JSON file",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "file",
				Usage:    "path to JSON file",
				Required: true,
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
			defer cancel()

			data, err := os.ReadFile(c.String("file"))
			if err != nil {
				return err
			}

			var newCourses []courses.NewCourse
			if err := json.Unmarshal(data, &newCourses); err != nil {
				return err
			}

			dto := courses.BulkWriteCoursesDTO{
				Courses: newCourses,
			}

			res, err := courses.BulkWriteCourses(ctx, db, dto)
			if err != nil {
				_ = printJSON(res)
				return fmt.Errorf("bulk write courses failed: %w", err)
			}
			return printJSON(res)
		},
	}
}

func createUserCommand(q usersdb.Querier) *cli.Command {
	return &cli.Command{
		Name:  "create-user",
		Usage: "create a new user",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "email",
				Required: true,
			},
			&cli.StringFlag{
				Name: "name",
			},
			&cli.Int64Flag{
				Name: "age",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
			defer cancel()

			dto := users.CreateUserDTO{
				Email: c.String("email"),
			}
			if c.IsSet("name") {
				name := c.String("name")
				dto.Name = &name
			}
			if c.IsSet("age") {
				age := c.Int64("age")
				dto.Age = &age
			}

			res, err := users.CreateUser(ctx, q, dto)
			if err != nil {
				return err
			}
			v := map[string]int64{"ID": res}

			return printJSON(v)
		},
	}
}

func updateUserCommand(q usersdb.Querier) *cli.Command {
	return &cli.Command{
		Name:  "update-user",
		Usage: "update user info",
		Flags: []cli.Flag{
			&cli.Int64Flag{
				Name: "id",
			},
			&cli.StringFlag{
				Name: "email",
			},
			&cli.StringFlag{
				Name: "name",
			},
			&cli.Int64Flag{
				Name: "age",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
			defer cancel()

			dto := users.UpdateUserDTO{
				ID: c.Int64("id"),
			}
			if c.IsSet("email") {
				email := c.String("email")
				dto.Email = &email
			}
			if c.IsSet("name") {
				name := c.String("name")
				dto.Name = &name
			}
			if c.IsSet("age") {
				age := c.Int64("age")
				dto.Age = &age
			}

			res, err := users.UpdateUser(ctx, q, dto)
			if err != nil {
				return err
			}
			return printJSON(res)
		},
	}
}
func findUserByIDCommand(q usersdb.Querier) *cli.Command {
	return &cli.Command{
		Name:  "find-user-by-id",
		Usage: "find a user by id (integer value)",
		Flags: []cli.Flag{
			&cli.Int64Flag{
				Name:     "id",
				Required: true,
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
			defer cancel()

			dto := users.FindUserByIDDTO{
				ID: c.Int64("id"),
			}

			res, err := users.FindUserByID(ctx, q, dto)
			if err != nil {
				return err
			}

			return printJSON(res)
		},
	}
}

func listUsersCommand(q usersdb.Querier) *cli.Command {
	return &cli.Command{
		Name:  "list-users",
		Usage: "list users",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "order",
				Value: "id_asc",
			},
			&cli.Int64Flag{
				Name:  "limit",
				Value: 10,
			},
			&cli.Int64Flag{
				Name:  "offset",
				Value: 0,
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
			defer cancel()

			dto := users.ListUsersDTO{
				OrderKey: c.String("order"),
				Limit:    c.Int64("limit"),
				Offset:   c.Int64("offset"),
			}

			res, err := users.ListUsers(ctx, q, dto)
			if err != nil {
				return err
			}

			return printJSON(res)
		},
	}
}

func deleteUserCommand(q usersdb.Querier) *cli.Command {
	return &cli.Command{
		Name:  "delete-user",
		Usage: "delete a user by id (integer value)",
		Flags: []cli.Flag{
			&cli.Int64Flag{
				Name:     "id",
				Required: true,
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
			defer cancel()

			dto := users.DeleteUserDTO{
				ID: c.Int64("id"),
			}

			err := users.DeleteUser(ctx, q, dto)
			if err != nil {
				return err
			}
			return nil
		},
	}
}

func parseInt64Strings(raw string) ([]int64, error) {
	parts := strings.Split(raw, ",")
	ids := make([]int64, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		id, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("invalid id %q: %w", part, err)
		}

		ids = append(ids, int64(id))
	}

	if len(ids) == 0 {
		return nil, fmt.Errorf("no valid ids provided")
	}

	return ids, nil
}

func printJSON(v any) error {
	payload, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(payload))
	return nil
}
