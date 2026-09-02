//go:build integration

package tasks_postgres_repository

import (
	"strings"
	"testing"
	"time"

	core_domain "github.com/Forestvov/checklist-service/internal/core/domain"
	core_pagination "github.com/Forestvov/checklist-service/internal/core/pagination"
)

func TestTasksRepositoryCreateTask(t *testing.T) {
	repository := newTestRepository(t)
	ctx := newTestContext(t)

	description := "Milk, bread and vegetables"
	priority := core_domain.TaskPriorityHigh
	dueAt := time.Date(2027, time.January, 10, 12, 0, 0, 0, time.UTC)
	input := core_domain.NewTaskUninitialized(
		"Buy groceries",
		&description,
		&priority,
		&dueAt,
	)

	created, err := repository.CreateTask(ctx, input)
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	if created.ID <= 0 {
		t.Fatalf("unexpected task ID: %d", created.ID)
	}
	if created.Version != initialTaskVersion {
		t.Errorf("unexpected initial version: got %d, want %d", created.Version, initialTaskVersion)
	}

	if created.Title != input.Title {
		t.Errorf("unexpected title: got %q, want %q", created.Title, input.Title)
	}
	if created.Description != input.Description {
		t.Errorf(
			"unexpected description: got %q, want %q",
			created.Description,
			input.Description,
		)
	}

	if created.Done {
		t.Error("new task must not be completed")
	}
	if created.Priority != priority {
		t.Errorf("unexpected priority: got %q, want %q", created.Priority, priority)
	}
	if created.DueAt == nil || !created.DueAt.Equal(dueAt) {
		t.Errorf("unexpected due_at: got %v, want %v", created.DueAt, dueAt)
	}

	if created.CreatedAt.IsZero() {
		t.Error("created_at must be set")
	}
	if created.UpdatedAt.IsZero() {
		t.Error("updated_at must be set")
	}

	stored, err := repository.GetTask(ctx, created.ID)
	if err != nil {
		t.Fatalf("get created task: %v", err)
	}

	if stored.ID != created.ID {
		t.Errorf("unexpected stored ID: got %d, want %d", stored.ID, created.ID)
	}
	if stored.Version != created.Version {
		t.Errorf("unexpected stored version: got %d, want %d", stored.Version, created.Version)
	}
	if stored.Title != created.Title {
		t.Errorf("unexpected stored title: got %q, want %q", stored.Title, created.Title)
	}
	if stored.Description != created.Description {
		t.Errorf(
			"unexpected stored description: got %q, want %q",
			stored.Description,
			created.Description,
		)
	}
	if stored.Done != created.Done {
		t.Errorf("unexpected stored done: got %t, want %t", stored.Done, created.Done)
	}
	if stored.Priority != created.Priority {
		t.Errorf("unexpected stored priority: got %q, want %q", stored.Priority, created.Priority)
	}
	if stored.DueAt == nil || !stored.DueAt.Equal(dueAt) {
		t.Errorf("unexpected stored due_at: got %v, want %v", stored.DueAt, dueAt)
	}
	if !stored.CreatedAt.Equal(created.CreatedAt) {
		t.Errorf(
			"unexpected stored created_at: got %s, want %s",
			stored.CreatedAt,
			created.CreatedAt,
		)
	}
	if !stored.UpdatedAt.Equal(created.UpdatedAt) {
		t.Errorf(
			"unexpected stored updated_at: got %s, want %s",
			stored.UpdatedAt,
			created.UpdatedAt,
		)
	}
}

func TestTasksRepositoryCreateTaskRejectsInvalidTitle(t *testing.T) {
	input := core_domain.NewTaskUninitialized("ab", nil, nil, nil)

	assertCreateTaskRejected(t, input)
}

func TestTasksRepositoryCreateTaskRejectsLongDescription(t *testing.T) {
	description := strings.Repeat("a", 5001)
	input := core_domain.NewTaskUninitialized("Valid title", &description, nil, nil)

	assertCreateTaskRejected(t, input)
}

func assertCreateTaskRejected(t *testing.T, input core_domain.Task) {
	t.Helper()

	repository := newTestRepository(t)
	ctx := newTestContext(t)

	created, err := repository.CreateTask(ctx, input)
	if err == nil {
		t.Fatal("expected database constraint error, got nil")
	}
	if created != (core_domain.Task{}) {
		t.Errorf("unexpected created task: got %+v, want zero task", created)
	}

	params, err := core_pagination.NewParams(nil, nil)
	if err != nil {
		t.Fatalf("create default pagination params: %v", err)
	}

	result, err := repository.GetTasks(
		ctx,
		params,
		core_domain.NewTaskFilter(nil, nil, nil, "", ""),
		time.Now().UTC(),
	)
	if err != nil {
		t.Fatalf("get tasks after rejected creation: %v", err)
	}
	if result.Total != 0 {
		t.Errorf("unexpected stored task total: got %d, want 0", result.Total)
	}
	if len(result.Items) != 0 {
		t.Errorf("unexpected stored task count: got %d, want 0", len(result.Items))
	}
}
