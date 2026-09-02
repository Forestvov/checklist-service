CREATE INDEX tasks_overdue_due_at_idx
	ON checklist.tasks (due_at)
	WHERE done = FALSE AND due_at IS NOT NULL;
