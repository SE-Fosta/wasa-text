CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(64) PRIMARY KEY,
    username VARCHAR(16) NOT NULL UNIQUE,
    photo_url TEXT
);

CREATE TABLE IF NOT EXISTS conversations (
    id VARCHAR(64) PRIMARY KEY,
    name VARCHAR(100), -- Nome del gruppo o nome dell'altra persona (cache)
    is_group BOOLEAN DEFAULT FALSE,
    photo_url TEXT,
    last_message_at DATETIME
);

CREATE TABLE IF NOT EXISTS participants (
    conversation_id VARCHAR(64),
    user_id VARCHAR(64),
    PRIMARY KEY (conversation_id, user_id),
    FOREIGN KEY (conversation_id) REFERENCES conversations(id),
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS messages (
    id VARCHAR(64) PRIMARY KEY,
    conversation_id VARCHAR(64),
    sender_id VARCHAR(64),
    content TEXT,
    photo_url TEXT, -- Se è un messaggio foto
    message_type VARCHAR(10) DEFAULT 'text', -- 'text' o 'photo'
    reply_to VARCHAR(64),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (conversation_id) REFERENCES conversations(id),
    FOREIGN KEY (sender_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS reactions (
    message_id VARCHAR(64),
    user_id VARCHAR(64),
    emoji VARCHAR(10),
    PRIMARY KEY (message_id, user_id),
    FOREIGN KEY (message_id) REFERENCES messages(id),
    FOREIGN KEY (user_id) REFERENCES users(id)
);
