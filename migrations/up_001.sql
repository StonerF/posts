CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS posts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    author_id VARCHAR(255) NOT NULL,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    allow_comments BOOLEAN NOT NULL
);

CREATE TABLE IF NOT EXISTS comments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    author_id VARCHAR(255) NOT NULL,
    post_id UUID NOT NULL,
    content TEXT NOT NULL,
    created_at timestamptz NOT NULL DEFAULT (now() AT TIME ZONE 'utc'),
    FOREIGN KEY (post_id) REFERENCES posts(id)
);

CREATE TABLE IF NOT EXISTS replies_comments (
    parent_comment_id UUID NOT NULL,
    reply_comment_id UUID NOT NULL,
    PRIMARY KEY (parent_comment_id, reply_comment_id),
    FOREIGN KEY (parent_comment_id) REFERENCES comments(id),
    FOREIGN KEY (reply_comment_id) REFERENCES comments(id)
);