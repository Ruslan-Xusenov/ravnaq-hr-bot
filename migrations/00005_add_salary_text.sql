-- +goose Up
ALTER TABLE vacancies ADD COLUMN salary_text VARCHAR(255);

-- +goose Down
ALTER TABLE vacancies DROP COLUMN salary_text;
