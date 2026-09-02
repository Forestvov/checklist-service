ALTER TABLE checklist.tasks
    ADD COLUMN version BIGINT NOT NULL DEFAULT 1,
    ADD CONSTRAINT tasks_version_positive_check
        CHECK ( version > 0 );