package tasks_postgres_repository

import "time"

type TaskModel struct {
	ID          int64
	Title       string
	Description string
	Done        bool
	CreatedAt   time.Time
	UpdateAt    time.Time
}
