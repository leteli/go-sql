package storage

import (
	"time"
)

type Course struct {
	ID    int64  `json:"id"`
	Slug  string `json:"slug"`
	Title string `json:"title"`
	Price int64  `json:"price"`
}

type User struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Name      *string   `json:"name,omitempty"`
	Age       *int64    `json:"age,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type Order struct {
	ID          int64     `json:"id"`
	UserID      *int64    `json:"user_id"`
	CourseID    *int64    `json:"course_id"`
	AmountCents int64     `json:"amount_cents"`
	CreatedAt   time.Time `json:"created_at"`
}

type CourseMember struct {
	ID       int64     `json:"id"`
	UserID   *int64    `json:"user_id"`
	CourseID *int64    `json:"course_id"`
	JoinedAt time.Time `json:"joined_at"`
}
