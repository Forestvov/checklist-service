ALTER TABLE checklist.tasks
    DROP CONSTRAINT tasks_version_positive_check,
    DROP COLUMN version;