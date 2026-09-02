//go:build integration

package tasks_postgres_repository

import (
	"errors"
	"testing"
	"time"

	core_domain "github.com/Forestvov/checklist-service/internal/core/domain"
	core_errors "github.com/Forestvov/checklist-service/internal/core/errors"
)

func TestTasksRepositoryUpdateTaskSuccess(t *testing.T) {
	repository := newTestRepository(t)
	ctx := newTestContext(t)

	created, err := repository.CreateTask(
		ctx,
		core_domain.NewTaskUninitialized("Original task", nil, nil, nil),
	)
	if err != nil {
		t.Fatalf("prepare task: %v", err)
	}

	description := "Updated description"
	done := true
	priority := core_domain.TaskPriorityHigh
	dueAt := created.CreatedAt.Add(24 * time.Hour)
	patch := core_domain.NewUpdateTask(
		core_domain.Nullable[string]{},
		setRepositoryNullable(description),
		setRepositoryNullable(done),
		setRepositoryNullable(priority),
		setRepositoryNullable(dueAt),
	)

	actual, err := repository.UpdateTask(
		ctx,
		created.ID,
		created.Version,
		patch,
	)
	if err != nil {
		t.Fatalf("update task: %v", err)
	}

	if actual.Title != created.Title {
		t.Errorf("title changed: got %q, want %q", actual.Title, created.Title)
	}
	if actual.Description != description {
		t.Errorf("unexpected description: got %q, want %q", actual.Description, description)
	}
	if actual.Done != done {
		t.Errorf("unexpected done: got %t, want %t", actual.Done, done)
	}
	if actual.Priority != priority {
		t.Errorf("unexpected priority: got %q, want %q", actual.Priority, priority)
	}
	if actual.DueAt == nil || !actual.DueAt.Equal(dueAt) {
		t.Errorf("unexpected due_at: got %v, want %v", actual.DueAt, dueAt)
	}
	if actual.Version != created.Version+1 {
		t.Errorf("unexpected version: got %d, want %d", actual.Version, created.Version+1)
	}
	if actual.UpdatedAt.Equal(created.UpdatedAt) {
		t.Errorf(
			"updated_at was not refreshed: before=%v after=%v",
			created.UpdatedAt,
			actual.UpdatedAt,
		)
	}

	stored, err := repository.GetTask(ctx, created.ID)
	if err != nil {
		t.Fatalf("get updated task: %v", err)
	}
	assertTasksEqual(t, stored, actual)
}

func TestTasksRepositoryUpdateTaskClearsDueAt(t *testing.T) {
	repository := newTestRepository(t)
	ctx := newTestContext(t)

	dueAt := time.Now().Add(24 * time.Hour)
	created, err := repository.CreateTask(
		ctx,
		core_domain.NewTaskUninitialized("Task with deadline", nil, nil, &dueAt),
	)
	if err != nil {
		t.Fatalf("prepare task: %v", err)
	}

	patch := core_domain.NewUpdateTask(
		core_domain.Nullable[string]{},
		core_domain.Nullable[string]{},
		core_domain.Nullable[bool]{},
		core_domain.Nullable[core_domain.TaskPriority]{},
		nullRepositoryNullable[time.Time](),
	)

	updated, err := repository.UpdateTask(
		ctx,
		created.ID,
		created.Version,
		patch,
	)
	if err != nil {
		t.Fatalf("clear due_at: %v", err)
	}

	if updated.DueAt != nil {
		t.Errorf("expected nil due_at, got %v", updated.DueAt)
	}
	if updated.Version != created.Version+1 {
		t.Errorf("unexpected version: got %d, want %d", updated.Version, created.Version+1)
	}
}

func TestTasksRepositoryUpdateTaskConflict(t *testing.T) {
	repository := newTestRepository(t)
	ctx := newTestContext(t)

	created, err := repository.CreateTask(
		ctx,
		core_domain.NewTaskUninitialized("Original task", nil, nil, nil),
	)
	if err != nil {
		t.Fatalf("prepare task: %v", err)
	}

	firstTitle := "First update"
	firstPatch := core_domain.NewUpdateTask(
		setRepositoryNullable(firstTitle),
		core_domain.Nullable[string]{},
		core_domain.Nullable[bool]{},
		core_domain.Nullable[core_domain.TaskPriority]{},
		core_domain.Nullable[time.Time]{},
	)

	firstUpdated, err := repository.UpdateTask(
		ctx,
		created.ID,
		created.Version,
		firstPatch,
	)
	if err != nil {
		t.Fatalf("first update: %v", err)
	}

	done := true
	stalePatch := core_domain.NewUpdateTask(
		core_domain.Nullable[string]{},
		core_domain.Nullable[string]{},
		setRepositoryNullable(done),
		core_domain.Nullable[core_domain.TaskPriority]{},
		core_domain.Nullable[time.Time]{},
	)

	actual, err := repository.UpdateTask(
		ctx,
		created.ID,
		created.Version,
		stalePatch,
	)
	if !errors.Is(err, core_errors.ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
	if actual != (core_domain.Task{}) {
		t.Errorf("expected empty task, got %+v", actual)
	}

	stored, err := repository.GetTask(ctx, created.ID)
	if err != nil {
		t.Fatalf("get stored task: %v", err)
	}
	if stored.Title != firstTitle {
		t.Errorf("unexpected title: got %q, want %q", stored.Title, firstTitle)
	}
	if stored.Done {
		t.Error("conflicting update must not modify done")
	}
	if stored.Version != firstUpdated.Version {
		t.Errorf(
			"version changed after conflict: got %d, want %d",
			stored.Version,
			firstUpdated.Version,
		)
	}
}

func TestTasksRepositoryUpdateTaskConcurrentConflict(t *testing.T) {
	repository := newTestRepository(t)
	ctx := newTestContext(t)

	created, err := repository.CreateTask(
		ctx,
		core_domain.NewTaskUninitialized("Original task", nil, nil, nil),
	)
	if err != nil {
		t.Fatalf("prepare task: %v", err)
	}

	updatedTitle := "Updated concurrently"
	patches := []core_domain.UpdateTask{
		core_domain.NewUpdateTask(
			setRepositoryNullable(updatedTitle),
			core_domain.Nullable[string]{},
			core_domain.Nullable[bool]{},
			core_domain.Nullable[core_domain.TaskPriority]{},
			core_domain.Nullable[time.Time]{},
		),
		core_domain.NewUpdateTask(
			core_domain.Nullable[string]{},
			core_domain.Nullable[string]{},
			setRepositoryNullable(true),
			core_domain.Nullable[core_domain.TaskPriority]{},
			core_domain.Nullable[time.Time]{},
		),
	}

	type updateResult struct {
		task core_domain.Task
		err  error
	}

	start := make(chan struct{})
	results := make(chan updateResult, len(patches))
	for _, patch := range patches {
		go func() {
			<-start
			task, updateErr := repository.UpdateTask(
				ctx,
				created.ID,
				created.Version,
				patch,
			)
			results <- updateResult{task: task, err: updateErr}
		}()
	}
	close(start)

	var successCount, conflictCount int
	for range patches {
		result := <-results
		switch {
		case result.err == nil:
			successCount++
			if result.task.Version != created.Version+1 {
				t.Errorf(
					"successful update version: got %d, want %d",
					result.task.Version,
					created.Version+1,
				)
			}
		case errors.Is(result.err, core_errors.ErrConflict):
			conflictCount++
		default:
			t.Errorf("unexpected concurrent update error: %v", result.err)
		}
	}

	if successCount != 1 || conflictCount != 1 {
		t.Fatalf(
			"unexpected results: successes=%d conflicts=%d",
			successCount,
			conflictCount,
		)
	}

	stored, err := repository.GetTask(ctx, created.ID)
	if err != nil {
		t.Fatalf("get stored task: %v", err)
	}
	if stored.Version != created.Version+1 {
		t.Errorf("stored version: got %d, want %d", stored.Version, created.Version+1)
	}
	if stored.Title == updatedTitle && stored.Done {
		t.Error("both concurrent patches were applied")
	}
	if stored.Title != updatedTitle && !stored.Done {
		t.Error("neither concurrent patch was applied")
	}
}

func TestTasksRepositoryUpdateTaskNotFound(t *testing.T) {
	const missingTaskID int64 = 9999

	title := "Updated task"
	patch := core_domain.NewUpdateTask(
		setRepositoryNullable(title),
		core_domain.Nullable[string]{},
		core_domain.Nullable[bool]{},
		core_domain.Nullable[core_domain.TaskPriority]{},
		core_domain.Nullable[time.Time]{},
	)

	repository := newTestRepository(t)
	ctx := newTestContext(t)

	actual, err := repository.UpdateTask(ctx, missingTaskID, initialTaskVersion, patch)
	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("unexpected error: got %v, want ErrNotFound", err)
	}
	if actual != (core_domain.Task{}) {
		t.Errorf("unexpected task: got %+v, want zero task", actual)
	}
}

func TestTasksRepositoryUpdateTaskRejectsInvalidTask(t *testing.T) {
	repository := newTestRepository(t)
	ctx := newTestContext(t)

	created, err := repository.CreateTask(
		ctx,
		core_domain.NewTaskUninitialized("Original task", nil, nil, nil),
	)
	if err != nil {
		t.Fatalf("prepare task: %v", err)
	}

	invalidTitle := "ab"
	patch := core_domain.NewUpdateTask(
		setRepositoryNullable(invalidTitle),
		core_domain.Nullable[string]{},
		core_domain.Nullable[bool]{},
		core_domain.Nullable[core_domain.TaskPriority]{},
		core_domain.Nullable[time.Time]{},
	)

	actual, err := repository.UpdateTask(ctx, created.ID, created.Version, patch)
	if err == nil {
		t.Fatal("expected database constraint error, got nil")
	}
	if actual != (core_domain.Task{}) {
		t.Errorf("unexpected task: got %+v, want zero task", actual)
	}

	stored, err := repository.GetTask(ctx, created.ID)
	if err != nil {
		t.Fatalf("get task after rejected update: %v", err)
	}
	assertTasksEqual(t, stored, created)
}

func assertTasksEqual(t *testing.T, actual, expected core_domain.Task) {
	t.Helper()

	if actual.ID != expected.ID {
		t.Errorf("unexpected ID: got %d, want %d", actual.ID, expected.ID)
	}
	if actual.Title != expected.Title {
		t.Errorf("unexpected title: got %q, want %q", actual.Title, expected.Title)
	}
	if actual.Description != expected.Description {
		t.Errorf("unexpected description: got %q, want %q", actual.Description, expected.Description)
	}
	if actual.Done != expected.Done {
		t.Errorf("unexpected done: got %t, want %t", actual.Done, expected.Done)
	}
	if actual.Priority != expected.Priority {
		t.Errorf("unexpected priority: got %q, want %q", actual.Priority, expected.Priority)
	}
	if actual.DueAt == nil && expected.DueAt != nil ||
		actual.DueAt != nil && expected.DueAt == nil ||
		actual.DueAt != nil && !actual.DueAt.Equal(*expected.DueAt) {
		t.Errorf("unexpected due_at: got %v, want %v", actual.DueAt, expected.DueAt)
	}
	if !actual.CreatedAt.Equal(expected.CreatedAt) {
		t.Errorf("unexpected created_at: got %s, want %s", actual.CreatedAt, expected.CreatedAt)
	}
	if !actual.UpdatedAt.Equal(expected.UpdatedAt) {
		t.Errorf("unexpected updated_at: got %s, want %s", actual.UpdatedAt, expected.UpdatedAt)
	}
	if actual.Version != expected.Version {
		t.Errorf("unexpected version: got %d, want %d", actual.Version, expected.Version)
	}
}

func setRepositoryNullable[T any](value T) core_domain.Nullable[T] {
	return core_domain.Nullable[T]{Value: &value, Set: true}
}

func nullRepositoryNullable[T any]() core_domain.Nullable[T] {
	return core_domain.Nullable[T]{Set: true}
}
