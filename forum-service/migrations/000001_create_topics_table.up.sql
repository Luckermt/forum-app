CREATE TABLE topics (
    id         VARCHAR(36) PRIMARY KEY,
    title      VARCHAR(100) NOT NULL,
    content    TEXT NOT NULL,
    user_id    VARCHAR(36) NOT NULL REFERENCES users(id),
    created_at TIMESTAMP NOT NULL,
    deleted    BOOLEAN NOT NULL DEFAULT FALSE
);