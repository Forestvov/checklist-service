package core_domain

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	core_errors "github.com/Forestvov/checklist-service/internal/core/errors"
)

const (
	minTaskTitleLength       = 3
	maxTaskTitleLength       = 255
	maxTaskDescriptionLength = 5000
)

type Task struct {
	ID          int64
	Title       string
	Description string
	Done        bool
	Priority    TaskPriority
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewTask(
	id int64,
	title string,
	description string,
	done bool,
	priority TaskPriority,
	createdAt time.Time,
	updatedAt time.Time,
) Task {
	return Task{
		ID:          id,
		Title:       title,
		Description: description,
		Done:        done,
		Priority:    priority,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}
}

func NewTaskUninitialized(
	title string,
	description *string,
	priority *TaskPriority,
) Task {
	descriptionValue := ""
	if description != nil {
		descriptionValue = *description
	}

	priorityValue := DefaultTaskPriority
	if priority != nil {
		priorityValue = *priority
	}

	now := time.Now()

	return NewTask(
		UninitializedID,
		title,
		descriptionValue,
		false,
		priorityValue,
		now,
		now,
	)
}

type UpdateTask struct {
	Title       Nullable[string]
	Description Nullable[string]
	Done        Nullable[bool]
	Priority    Nullable[TaskPriority]
}

func NewUpdateTask(
	title Nullable[string],
	description Nullable[string],
	done Nullable[bool],
	priority Nullable[TaskPriority],
) UpdateTask {
	return UpdateTask{
		Title:       title,
		Description: description,
		Done:        done,
		Priority:    priority,
	}
}

func (u UpdateTask) Validate() error {
	if !u.Title.Set &&
		!u.Description.Set &&
		!u.Done.Set &&
		!u.Priority.Set {
		return fmt.Errorf("at least one task field must be provided: %w", core_errors.ErrInvalidArgument)
	}

	if u.Title.Set {
		if u.Title.Value == nil {
			return fmt.Errorf("title must not be null: %w", core_errors.ErrInvalidArgument)
		}

		if err := validateTaskTitle(*u.Title.Value); err != nil {
			return err
		}
	}

	if u.Description.Set {
		if u.Description.Value == nil {
			return fmt.Errorf("description must not be null: %w", core_errors.ErrInvalidArgument)
		}

		if err := validateTaskDescription(*u.Description.Value); err != nil {
			return err
		}
	}

	if u.Done.Set && u.Done.Value == nil {
		return fmt.Errorf("done must not be null: %w", core_errors.ErrInvalidArgument)
	}

	if u.Priority.Set {
		if u.Priority.Value == nil {
			return fmt.Errorf(
				"priority must not be null: %w",
				core_errors.ErrInvalidArgument,
			)
		}

		if err := u.Priority.Value.Validate(); err != nil {
			return fmt.Errorf("validate priority: %w", err)
		}
	}

	return nil
}

func (t *Task) Validate() error {
	if err := validateTaskTitle(t.Title); err != nil {
		return err
	}

	if err := validateTaskDescription(t.Description); err != nil {
		return err
	}

	if err := t.Priority.Validate(); err != nil {
		return fmt.Errorf("validate task priority: %w", err)
	}

	return nil
}

func validateTaskTitle(titleValue string) error {
	title := strings.TrimSpace(titleValue)
	titleLen := utf8.RuneCountInString(title)

	if titleLen < minTaskTitleLength || titleLen > maxTaskTitleLength {
		return fmt.Errorf(
			"invalid title length: %d, must be between %d and %d: %w",
			titleLen,
			minTaskTitleLength,
			maxTaskTitleLength,
			core_errors.ErrInvalidArgument,
		)
	}

	return nil
}

func validateTaskDescription(description string) error {
	descriptionLen := utf8.RuneCountInString(description)
	if descriptionLen > maxTaskDescriptionLength {
		return fmt.Errorf(
			"invalid description length: %d, must not exceed %d: %w",
			descriptionLen,
			maxTaskDescriptionLength,
			core_errors.ErrInvalidArgument,
		)
	}

	return nil
}

func (t *Task) ApplyUpdate(patch UpdateTask) error {
	if err := patch.Validate(); err != nil {
		return fmt.Errorf("validate task patch: %w", err)
	}

	tmp := *t

	if patch.Title.Set {
		tmp.Title = *patch.Title.Value
	}

	if patch.Description.Set {
		tmp.Description = *patch.Description.Value
	}

	if patch.Done.Set {
		tmp.Done = *patch.Done.Value
	}

	if patch.Priority.Set {
		tmp.Priority = *patch.Priority.Value
	}

	if err := tmp.Validate(); err != nil {
		return fmt.Errorf("validate patched task: %w", err)
	}

	tmp.UpdatedAt = time.Now()
	*t = tmp

	return nil
}
