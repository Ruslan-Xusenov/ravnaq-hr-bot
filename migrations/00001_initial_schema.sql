-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    telegram_id BIGINT UNIQUE NOT NULL,
    telegram_username VARCHAR(255),
    telegram_first_name VARCHAR(255),
    telegram_last_name VARCHAR(255),
    language_code VARCHAR(5),
    primary_phone VARCHAR(50),
    status VARCHAR(50) DEFAULT 'active',
    last_activity_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE TABLE resumes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    version INT NOT NULL DEFAULT 1,
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    photo_file_id UUID,
    address_region VARCHAR(255),
    address_city VARCHAR(255),
    address_detail TEXT,
    expected_salary NUMERIC,
    salary_currency VARCHAR(10),
    education_text TEXT,
    skills_text TEXT,
    languages_text TEXT,
    portfolio_url TEXT,
    pdf_file_id UUID,
    consent_at TIMESTAMP,
    is_current BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE resume_phones (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    resume_id UUID NOT NULL REFERENCES resumes(id) ON DELETE CASCADE,
    phone VARCHAR(50) NOT NULL,
    position SMALLINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE work_experiences (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    resume_id UUID NOT NULL REFERENCES resumes(id) ON DELETE CASCADE,
    company_name VARCHAR(255) NOT NULL,
    position_name VARCHAR(255) NOT NULL,
    started_at DATE NOT NULL,
    ended_at DATE,
    is_current BOOLEAN DEFAULT false,
    duties TEXT NOT NULL,
    leaving_reason TEXT,
    city VARCHAR(255),
    sort_order INT NOT NULL DEFAULT 0
);

CREATE TABLE vacancies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    department VARCHAR(255),
    location VARCHAR(255),
    employment_type VARCHAR(100),
    schedule VARCHAR(100),
    salary_from NUMERIC,
    salary_to NUMERIC,
    salary_currency VARCHAR(10),
    description TEXT,
    requirements TEXT,
    benefits TEXT,
    external_url TEXT,
    status VARCHAR(50) DEFAULT 'draft',
    published_at TIMESTAMP,
    expires_at TIMESTAMP,
    sort_order INT DEFAULT 0,
    created_by UUID,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE applications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    vacancy_id UUID NOT NULL REFERENCES vacancies(id) ON DELETE CASCADE,
    resume_id UUID NOT NULL REFERENCES resumes(id),
    status VARCHAR(50) DEFAULT 'submitted',
    admin_note TEXT,
    rejection_reason TEXT,
    assigned_admin_id UUID,
    interview_at TIMESTAMP,
    accepted_at TIMESTAMP,
    link_sent_at TIMESTAMP,
    submitted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, vacancy_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS applications;
DROP TABLE IF EXISTS vacancies;
DROP TABLE IF EXISTS work_experiences;
DROP TABLE IF EXISTS resume_phones;
DROP TABLE IF EXISTS resumes;
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
