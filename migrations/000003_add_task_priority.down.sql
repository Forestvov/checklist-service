ALTER TABLE checklist.tasks
    DROP CONSTRAINT tasks_priority_check,
    DROP COLUMN priority;
