package tasks_transport_http

import (
	"time"

	core_domain "github.com/Forestvov/checklist-service/internal/core/domain"
)

type TaskDTOResponse struct {
	ID          int64                    `json:"id"`
	Title       string                   `json:"title"`
	Description string                   `json:"description"`
	Done        bool                     `json:"done"`
	Priority    core_domain.TaskPriority `json:"priority" enums:"low,medium,high"`
	CreatedAt   time.Time                `json:"created_at"`
	UpdatedAt   time.Time                `json:"updated_at"`
	DueAt       *time.Time               `json:"due_at" format:"date-time"`
}

func taskDTOFromDomain(task core_domain.Task) TaskDTOResponse {
	return TaskDTOResponse{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Done:        task.Done,
		Priority:    task.Priority,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
		DueAt:       task.DueAt,
	}
}

func tasksDTOsFromDomain(tasks []core_domain.Task) []TaskDTOResponse {
	dtos := make([]TaskDTOResponse, len(tasks))

	for i, task := range tasks {
		dtos[i] = taskDTOFromDomain(task)
	}

	return dtos
}
