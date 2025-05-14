CREATE TABLE messages (
    id         VARCHAR(36) PRIMARY KEY,
    topic_id   VARCHAR(36) REFERENCES topics(id),
    user_id    VARCHAR(36) NOT NULL REFERENCES users(id),
    content    TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    is_chat    BOOLEAN NOT NULL DEFAULT FALSE
);