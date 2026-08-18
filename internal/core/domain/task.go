package core_domain

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	core_errors "github.com/Forestvov/checklist-service/internal/core/errors"
)

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

func NewTaskUninitialized(
	title string,
	description *string,
) Task {
	descriptionValue := ""
	if description != nil {
		descriptionValue = *description
	}

	return NewTask(
		int64(UninitializedID),
		title,
		descriptionValue,
		false,
		time.Now(),
		time.Now(),
	)
}

func (t *Task) Validate() error {
	title := strings.TrimSpace(t.Title)
	titleLen := utf8.RuneCountInString(title)

	if titleLen < 3 || titleLen > 255 {
		return fmt.Errorf(
			"invalid title length: %d, must be between 3 and 255: %w",
			titleLen,
			core_errors.ErrInvalidArgument,
		)
	}

	descriptionLen := utf8.RuneCountInString(t.Description)
	if descriptionLen > 5000 {
		return fmt.Errorf(
			"invalid description length: %d, must not exceed 5000: %w",
			descriptionLen,
			core_errors.ErrInvalidArgument,
		)
	}

	return nil
}
