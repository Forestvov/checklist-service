ALTER TABLE checklist.tasks
    ADD COLUMN priority VARCHAR(16) NOT NULL DEFAULT 'medium',
    ADD CONSTRAINT tasks_priority_check
        CHECK (priority IN ('low', 'medium', 'high'));
