package tasks_postgres_repository

import (
	"time"

	core_domain "github.com/Forestvov/checklist-service/internal/core/domain"
)

type TaskModel struct {
	ID          int64
	Title       string
	Description string
	Done        bool
	CreatedAt   time.Time
	UpdateAt    time.Time
}

func taskDomainFromModel(taskModel TaskModel) core_domain.Task {
	return core_domain.NewTask(
		taskModel.ID,
		taskModel.Title,
		taskModel.Description,
		taskModel.Done,
		taskModel.CreatedAt,
		taskModel.UpdateAt,
	)
}

func taskDomainFromModels(taskModels []TaskModel) []core_domain.Task {
	domains := make([]core_domain.Task, len(taskModels))

	for i, model := range taskModels {
		domains[i] = taskDomainFromModel(model)
	}

	return domains
}
