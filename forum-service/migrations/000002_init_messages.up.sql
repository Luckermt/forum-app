CREATE TABLE messages (
    id UUID PRIMARY KEY,
    topic_id UUID REFERENCES topics(id),
    user_id UUID NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    is_chat BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX idx_messages_topic_id ON messages(topic_id);
CREATE INDEX idx_messages_user_id ON messages(user_id);
CREATE INDEX idx_messages_created_at ON messages(created_at);