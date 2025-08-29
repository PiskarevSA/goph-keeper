-- +goose Up
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL
);

CREATE TABLE secrets (
    uuid UUID PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    kind INT NOT NULL,
    created TIMESTAMP NOT NULL,
    modified TIMESTAMP NOT NULL,
    description TEXT
);

CREATE TABLE credentials_secrets (
    uuid UUID PRIMARY KEY REFERENCES secrets(uuid) ON DELETE CASCADE,
    login TEXT NOT NULL,
    password TEXT NOT NULL
);

CREATE TABLE card_secrets (
    uuid UUID PRIMARY KEY REFERENCES secrets(uuid) ON DELETE CASCADE,
    number TEXT NOT NULL,
    holder TEXT NOT NULL,
    expires TEXT NOT NULL,
    verification_code TEXT NOT NULL
);

CREATE TABLE text_secrets (
    uuid UUID PRIMARY KEY REFERENCES secrets(uuid) ON DELETE CASCADE,
    filename TEXT NOT NULL,
    content TEXT NOT NULL
);

CREATE TABLE raw_secrets (
    uuid UUID PRIMARY KEY REFERENCES secrets(uuid) ON DELETE CASCADE,
    filename TEXT NOT NULL,
    path TEXT NOT NULL,
    size BIGINT NOT NULL
);

-- +goose Down
DROP TABLE raw_secrets;
DROP TABLE text_secrets;
DROP TABLE card_secrets;
DROP TABLE credentials_secrets;
DROP TABLE secrets;
DROP TABLE users;
