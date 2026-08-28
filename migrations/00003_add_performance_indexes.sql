-- +goose Up
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_resumes_user_id ON resumes(user_id);
CREATE INDEX IF NOT EXISTS idx_applications_user_id ON applications(user_id);
CREATE INDEX IF NOT EXISTS idx_applications_vacancy_id ON applications(vacancy_id);
CREATE INDEX IF NOT EXISTS idx_vacancies_status_published_at ON vacancies(status, published_at DESC);
CREATE INDEX IF NOT EXISTS idx_vacancies_created_at ON vacancies(created_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_vacancies_created_at;
DROP INDEX IF EXISTS idx_vacancies_status_published_at;
DROP INDEX IF EXISTS idx_applications_vacancy_id;
DROP INDEX IF EXISTS idx_applications_user_id;
DROP INDEX IF EXISTS idx_resumes_user_id;
-- +goose StatementEnd
