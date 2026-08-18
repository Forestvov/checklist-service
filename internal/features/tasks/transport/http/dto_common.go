package tasks_transport_http

import (
	"time"

	core_domain "github.com/Forestvov/checklist-service/internal/core/domain"
)

type TaskDTOResponse struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Done        bool      `json:"done"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func taskDTOFromDomain(task core_domain.Task) TaskDTOResponse {
	return TaskDTOResponse{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Done:        task.Done,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdateAt,
	}
}
