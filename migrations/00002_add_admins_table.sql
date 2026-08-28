-- +goose Up
-- +goose StatementBegin
CREATE TABLE admins (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(50) DEFAULT 'admin',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO admins (email, password, role)
VALUES ('admin@example.com', '$2a$10$XZO0eZEbAJHfbp4d/S4Ky.2fw4QjR5rHp.2Xy.z39ugsrYoNPKRiO', 'superadmin');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS admins;
-- +goose StatementEnd
