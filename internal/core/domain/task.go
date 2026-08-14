package core_domain

import "time"

type Task struct {
	ID          int64
	Title       string
	Description string
	Done        bool
	CreatedAt   time.Time
	UpdateAt    time.Time
}

func NewTask(
	id int64,
	title string,
	description string,
	done bool,
	createdAt time.Time,
	updateAt time.Time,
) Task {
	return Task{
		ID:          id,
		Title:       title,
		Description: description,
		Done:        done,
		CreatedAt:   createdAt,
		UpdateAt:    updateAt,
	}
}
