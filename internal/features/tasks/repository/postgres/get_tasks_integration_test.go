//go:build integration

package tasks_postgres_repository

import (
	"slices"
	"testing"
	"time"

	core_domain "github.com/Forestvov/checklist-service/internal/core/domain"
	core_pagination "github.com/Forestvov/checklist-service/internal/core/pagination"
)

func TestTasksRepositoryGetTasksSorting(t *testing.T) {
	repository := newTestRepository(t)
	ctx := newTestContext(t)
	baseTime := time.Date(2026, time.August, 30, 12, 0, 0, 0, time.UTC)

	inputs := []core_domain.Task{
		core_domain.NewTask(
			core_domain.UninitializedID,
			"Charlie task",
			"",
			false,
			baseTime.Add(time.Minute),
			baseTime.Add(3*time.Minute),
		),
		core_domain.NewTask(
			core_domain.UninitializedID,
			"Alpha task",
			"",
			false,
			baseTime.Add(2*time.Minute),
			baseTime.Add(time.Minute),
		),
		core_domain.NewTask(
			core_domain.UninitializedID,
			"Bravo task",
			"",
			false,
			baseTime.Add(3*time.Minute),
			baseTime.Add(2*time.Minute),
		),
	}

	createdTasks := make([]core_domain.Task, 0, len(inputs))
	for _, input := range inputs {
		created, err := repository.CreateTask(ctx, input)
		if err != nil {
			t.Fatalf("prepare task %q: %v", input.Title, err)
		}
		createdTasks = append(createdTasks, created)
	}

	params, err := core_pagination.NewParams(nil, nil)
	if err != nil {
		t.Fatalf("create default pagination params: %v", err)
	}

	tests := []struct {
		name        string
		sort        core_domain.TaskSort
		order       core_domain.SortOrder
		expectedIDs []int64
	}{
		{
			name:        "created at ascending",
			sort:        core_domain.TaskSortCreatedAt,
			order:       core_domain.SortOrderAsc,
			expectedIDs: []int64{createdTasks[0].ID, createdTasks[1].ID, createdTasks[2].ID},
		},
		{
			name:        "created at descending",
			sort:        core_domain.TaskSortCreatedAt,
			order:       core_domain.SortOrderDesc,
			expectedIDs: []int64{createdTasks[2].ID, createdTasks[1].ID, createdTasks[0].ID},
		},
		{
			name:        "updated at ascending",
			sort:        core_domain.TaskSortUpdatedAt,
			order:       core_domain.SortOrderAsc,
			expectedIDs: []int64{createdTasks[1].ID, createdTasks[2].ID, createdTasks[0].ID},
		},
		{
			name:        "updated at descending",
			sort:        core_domain.TaskSortUpdatedAt,
			order:       core_domain.SortOrderDesc,
			expectedIDs: []int64{createdTasks[0].ID, createdTasks[2].ID, createdTasks[1].ID},
		},
		{
			name:        "title ascending",
			sort:        core_domain.TaskSortTitle,
			order:       core_domain.SortOrderAsc,
			expectedIDs: []int64{createdTasks[1].ID, createdTasks[2].ID, createdTasks[0].ID},
		},
		{
			name:        "title descending",
			sort:        core_domain.TaskSortTitle,
			order:       core_domain.SortOrderDesc,
			expectedIDs: []int64{createdTasks[0].ID, createdTasks[2].ID, createdTasks[1].ID},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repository.GetTasks(
				ctx,
				params,
				core_domain.NewTaskFilter(nil, tt.sort, tt.order),
			)
			if err != nil {
				t.Fatalf("get sorted tasks: %v", err)
			}

			actualIDs := make([]int64, 0, len(result.Items))
			for _, task := range result.Items {
				actualIDs = append(actualIDs, task.ID)
			}
			if !slices.Equal(actualIDs, tt.expectedIDs) {
				t.Errorf("unexpected task order: got %v, want %v", actualIDs, tt.expectedIDs)
			}
		})
	}
}

func TestTasksRepositoryGetTasksDoneFilter(t *testing.T) {
	repository := newTestRepository(t)
	ctx := newTestContext(t)

	createdTasks := make([]core_domain.Task, 0, 5)
	for _, title := range []string{
		"First task",
		"Second task",
		"Third task",
		"Fourth task",
		"Fifth task",
	} {
		created, err := repository.CreateTask(
			ctx,
			core_domain.NewTaskUninitialized(title, nil),
		)
		if err != nil {
			t.Fatalf("prepare task %q: %v", title, err)
		}
		createdTasks = append(createdTasks, created)
	}

	for _, index := range []int{1, 3} {
		completed := createdTasks[index]
		completed.Done = true
		completed.UpdatedAt = completed.UpdatedAt.Add(time.Second)

		updated, err := repository.UpdateTask(ctx, completed.ID, completed)
		if err != nil {
			t.Fatalf("complete task %d: %v", completed.ID, err)
		}
		createdTasks[index] = updated
	}

	params, err := core_pagination.NewParams(nil, nil)
	if err != nil {
		t.Fatalf("create default pagination params: %v", err)
	}

	tests := []struct {
		name        string
		done        bool
		expectedIDs []int64
	}{
		{
			name:        "completed tasks",
			done:        true,
			expectedIDs: []int64{createdTasks[3].ID, createdTasks[1].ID},
		},
		{
			name:        "uncompleted tasks",
			done:        false,
			expectedIDs: []int64{createdTasks[4].ID, createdTasks[2].ID, createdTasks[0].ID},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repository.GetTasks(
				ctx,
				params,
				core_domain.NewTaskFilter(&tt.done, "", ""),
			)
			if err != nil {
				t.Fatalf("get filtered tasks: %v", err)
			}
			if result.Total != int64(len(tt.expectedIDs)) {
				t.Errorf("unexpected total: got %d, want %d", result.Total, len(tt.expectedIDs))
			}
			if len(result.Items) != len(tt.expectedIDs) {
				t.Fatalf("unexpected item count: got %d, want %d", len(result.Items), len(tt.expectedIDs))
			}
			for index, task := range result.Items {
				if task.ID != tt.expectedIDs[index] {
					t.Errorf("item %d: unexpected ID: got %d, want %d", index, task.ID, tt.expectedIDs[index])
				}
				if task.Done != tt.done {
					t.Errorf("item %d: unexpected done: got %t, want %t", index, task.Done, tt.done)
				}
			}
		})
	}

	page := 2
	perPage := 2
	filteredParams, err := core_pagination.NewParams(&page, &perPage)
	if err != nil {
		t.Fatalf("create filtered pagination params: %v", err)
	}
	done := false
	result, err := repository.GetTasks(
		ctx,
		filteredParams,
		core_domain.NewTaskFilter(&done, "", ""),
	)
	if err != nil {
		t.Fatalf("get filtered task page: %v", err)
	}
	if result.Total != 3 {
		t.Errorf("unexpected filtered total: got %d, want 3", result.Total)
	}
	if result.TotalPages() != 2 {
		t.Errorf("unexpected filtered total pages: got %d, want 2", result.TotalPages())
	}
	if len(result.Items) != 1 || result.Items[0].ID != createdTasks[0].ID {
		t.Errorf("unexpected second filtered page: got %+v", result.Items)
	}
}

func TestTasksRepositoryGetTasksEmpty(t *testing.T) {
	repository := newTestRepository(t)
	ctx := newTestContext(t)

	params, err := core_pagination.NewParams(nil, nil)
	if err != nil {
		t.Fatalf("create default pagination params: %v", err)
	}

	result, err := repository.GetTasks(
		ctx,
		params,
		core_domain.NewTaskFilter(nil, "", ""),
	)
	if err != nil {
		t.Fatalf("get tasks from migrated database: %v", err)
	}
	if result.Total != 0 {
		t.Fatalf("unexpected total tasks: got %d, want 0", result.Total)
	}
	if result.Items == nil {
		t.Fatal("unexpected nil items, want empty slice")
	}
	if len(result.Items) != 0 {
		t.Fatalf("unexpected task count: got %d, want 0", len(result.Items))
	}
}

func TestTasksRepositoryGetTasksSuccess(t *testing.T) {
	repository := newTestRepository(t)
	ctx := newTestContext(t)

	titles := []string{
		"First task",
		"Second task",
		"Third task",
	}

	createdTasks := make([]core_domain.Task, 0, len(titles))

	for _, title := range titles {
		input := core_domain.NewTaskUninitialized(title, nil)

		created, err := repository.CreateTask(ctx, input)
		if err != nil {
			t.Fatalf("prepare task %q: %v", title, err)
		}

		createdTasks = append(createdTasks, created)
	}

	params, err := core_pagination.NewParams(nil, nil)
	if err != nil {
		t.Fatalf("create default pagination params: %v", err)
	}

	result, err := repository.GetTasks(
		ctx,
		params,
		core_domain.NewTaskFilter(nil, "", ""),
	)
	if err != nil {
		t.Fatalf("get tasks: %v", err)
	}

	if result.Total != int64(len(createdTasks)) {
		t.Errorf(
			"unexpected total: got %d, want %d",
			result.Total,
			len(createdTasks),
		)
	}

	if len(result.Items) != len(createdTasks) {
		t.Fatalf(
			"unexpected task count: got %d, want %d",
			len(result.Items),
			len(createdTasks),
		)
	}

	for i, actual := range result.Items {
		expected := createdTasks[len(createdTasks)-1-i]

		if actual.ID != expected.ID {
			t.Errorf(
				"item %d: unexpected ID: got %d, want %d",
				i,
				actual.ID,
				expected.ID,
			)
		}

		if actual.Title != expected.Title {
			t.Errorf(
				"item %d: unexpected title: got %q, want %q",
				i,
				actual.Title,
				expected.Title,
			)
		}
	}

	if result.Params != params {
		t.Errorf(
			"unexpected params: got %+v, want %+v",
			result.Params,
			params,
		)
	}

	if result.TotalPages() != 1 {
		t.Errorf(
			"unexpected total pages: got %d, want 1",
			result.TotalPages(),
		)
	}
}

func TestTasksRepositoryGetTasksPagination(t *testing.T) {
	repository := newTestRepository(t)
	ctx := newTestContext(t)

	titles := []string{
		"First task",
		"Second task",
		"Third task",
		"Fourth task",
		"Fifth task",
	}

	createdTasks := make([]core_domain.Task, 0, len(titles))

	for _, title := range titles {
		input := core_domain.NewTaskUninitialized(title, nil)

		created, err := repository.CreateTask(ctx, input)
		if err != nil {
			t.Fatalf("prepare task %q: %v", title, err)
		}

		createdTasks = append(createdTasks, created)
	}

	page := 2
	perPage := 2

	params, err := core_pagination.NewParams(&page, &perPage)
	if err != nil {
		t.Fatalf("create pagination params: %v", err)
	}

	result, err := repository.GetTasks(
		ctx,
		params,
		core_domain.NewTaskFilter(nil, "", ""),
	)
	if err != nil {
		t.Fatalf("get tasks: %v", err)
	}

	expectedTotal := int64(len(createdTasks))
	if result.Total != expectedTotal {
		t.Errorf(
			"unexpected total: got %d, want %d",
			result.Total,
			expectedTotal,
		)
	}

	if len(result.Items) != perPage {
		t.Fatalf(
			"unexpected task count: got %d, want %d",
			len(result.Items),
			perPage,
		)
	}

	const expectedTotalPages int64 = 3
	if result.TotalPages() != expectedTotalPages {
		t.Errorf(
			"unexpected total pages: got %d, want %d",
			result.TotalPages(),
			expectedTotalPages,
		)
	}

	if result.Params != params {
		t.Errorf(
			"unexpected params: got %+v, want %+v",
			result.Params,
			params,
		)
	}

	expectedTasks := []core_domain.Task{
		createdTasks[2],
		createdTasks[1],
	}

	for i, actual := range result.Items {
		expected := expectedTasks[i]

		if actual.ID != expected.ID {
			t.Errorf(
				"item %d: unexpected ID: got %d, want %d",
				i,
				actual.ID,
				expected.ID,
			)
		}

		if actual.Title != expected.Title {
			t.Errorf(
				"item %d: unexpected title: got %q, want %q",
				i,
				actual.Title,
				expected.Title,
			)
		}
	}
}
