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
	"strconv"

	"encoding/json"

	"github.com/urfave/cli/v3"
	_ "modernc.org/sqlite"
)

type User struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	db, err := sql.Open("sqlite", "file:data.db?_foreign_keys=on&_busy_timeout=5000")
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("ping db: %v", err)
	}

	cmd := &cli.Command{
		Name:  "db-tool",
		Usage: "CLI for testing course DB operations",
		Commands: []*cli.Command{
			createCourseCommand(db),
			listCoursesCommand(db),
			findCoursesByIDsCommand(db),
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

func createCourseCommand(db *sql.DB) *cli.Command {
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
			&cli.IntFlag{
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
				Price: c.Int("price"),
			}

			res, err := courses.CreateCourse(ctx, db, dto)
			if err != nil {
				return err
			}

			return printJSON(res)
		},
	}
}

func listCoursesCommand(db *sql.DB) *cli.Command {
	return &cli.Command{
		Name:  "list-courses",
		Usage: "list courses",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "order",
				Value: "id_asc",
			},
			&cli.IntFlag{
				Name:  "limit",
				Value: 10,
			},
			&cli.IntFlag{
				Name:  "offset",
				Value: 0,
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
			defer cancel()

			dto := courses.ListCoursesDTO{
				OrderKey: c.String("order"),
				Limit:    c.Int("limit"),
				Offset:   c.Int("offset"),
			}

			res, err := courses.ListCourses(ctx, db, dto)
			if err != nil {
				return err
			}

			return printJSON(res)
		},
	}
}

func findCoursesByIDsCommand(db *sql.DB) *cli.Command {
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

			ids, err := parseIDs(c.String("ids"))
			if err != nil {
				return err
			}
			dto := courses.FindCoursesByIDsDTO{
				IDs: ids,
			}

			res, err := courses.FindCoursesByIDs(ctx, db, dto)
			if err != nil {
				return err
			}

			return printJSON(res)
		},
	}
}

func parseIDs(raw string) ([]int, error) {
	parts := strings.Split(raw, ",")
	ids := make([]int, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		id, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("invalid id %q: %w", part, err)
		}

		ids = append(ids, id)
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
