-- +goose Up
-- +goose StatementBegin
ALTER TABLE resumes ALTER COLUMN photo_file_id TYPE VARCHAR(255);
ALTER TABLE resumes ADD COLUMN extra_phone_1 VARCHAR(50);
ALTER TABLE resumes ADD COLUMN extra_phone_2 VARCHAR(50);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE resumes DROP COLUMN extra_phone_2;
ALTER TABLE resumes DROP COLUMN extra_phone_1;
-- Note: Reverting photo_file_id back to UUID might fail if string contains non-uuid values,
-- so down migration for that column is risky. We'll cast it to UUID and hope for the best.
ALTER TABLE resumes ALTER COLUMN photo_file_id TYPE UUID USING photo_file_id::uuid;
-- +goose StatementEnd
