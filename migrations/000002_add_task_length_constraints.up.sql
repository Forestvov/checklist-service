ALTER TABLE checklist.tasks
    ADD CONSTRAINT tasks_title_length_check
        CHECK (char_length(btrim(title)) BETWEEN 3 AND 255),
    ADD CONSTRAINT tasks_description_length_check
        CHECK (char_length(description) <= 5000);
