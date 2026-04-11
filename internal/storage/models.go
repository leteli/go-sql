package storage

import (
	"time"
)

type Course struct {
	ID    int    `json:"id"`
	Slug  string `json:"slug"`
	Title string `json:"title"`
	Price int    `json:"price"`
}

type User struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	Name      *string   `json:"name,omitempty"`
	Age       *int      `json:"age,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}
