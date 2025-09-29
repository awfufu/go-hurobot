CREATE TABLE users (
    "user_id"     BIGINT NOT NULL,
    "name"        TEXT NOT NULL,
    "nick_name"   TEXT,
    "summary"     TEXT,
    "token_usage" BIGINT NOT NULL DEFAULT 0,
    PRIMARY KEY ("user_id")
);

CREATE TABLE suppliers (
    "name"          TEXT NOT NULL,
    "base_url"      TEXT NOT NULL,
    "api_key"       TEXT,
    "default_model" TEXT,
    PRIMARY KEY ("name")
);

CREATE TABLE messages (
    "msg_id"   BIGINT NOT NULL,
    "user_id"  BIGINT NOT NULL,
    "group_id" BIGINT NOT NULL,
    "content"  TEXT NOT NULL,
    "raw"      TEXT,
    "deleted"  BOOLEAN NOT NULL DEFAULT FALSE,
    "is_cmd"   BOOLEAN NOT NULL DEFAULT FALSE,
    "time"     TIMESTAMP WITH TIME ZONE NOT NULL,
    PRIMARY KEY ("msg_id"),
    FOREIGN KEY ("user_id") REFERENCES users(user_id)
);

CREATE TABLE group_llm_configs (
    "group_id"    BIGINT  NOT NULL,
    "prompt"      TEXT    NOT NULL,
    "max_history" INTEGER NOT NULL DEFAULT 100,
    "enabled"     BOOLEAN NOT NULL DEFAULT TRUE,
    "info"        TEXT,
    "debug"       BOOLEAN NOT NULL DEFAULT FALSE,
    "supplier"    TEXT,
    "model"       TEXT,
    PRIMARY KEY ("group_id"),
    FOREIGN KEY ("supplier") REFERENCES suppliers(name)
);

CREATE INDEX idx_messages_covering ON messages("group_id", "is_cmd", "time" DESC, "user_id", "content", "msg_id");

CREATE TABLE user_events (
    "user_id"    BIGINT NOT NULL,
    "event_idx"  INTEGER NOT NULL,
    "msg_regex"  TEXT NOT NULL,
    "reply_text" TEXT NOT NULL,
    "rand_prob"  REAL NOT NULL DEFAULT 1.0,
    "created_at" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    PRIMARY KEY ("user_id", "event_idx"),
    FOREIGN KEY ("user_id") REFERENCES users(user_id),
    CONSTRAINT check_rand_prob CHECK (rand_prob >= 0.0 AND rand_prob <= 1.0),
    CONSTRAINT check_event_idx CHECK (event_idx >= 0 AND event_idx <= 9)
);

CREATE TABLE group_rcon_configs (
    "group_id" BIGINT NOT NULL,
    "address"  TEXT NOT NULL,
    "password" TEXT NOT NULL,
    "enabled"  BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY ("group_id")
);

INSERT INTO suppliers ("name", "base_url", "api_key", "default_model") VALUES
('siliconflow', 'https://api.siliconflow.cn/v1', '', 'deepseek-ai/DeepSeek-V3.1');

CREATE TABLE legacy_game (
    "user_id" BIGINT NOT NULL,
    "energy"  INT NOT NULL DEFAULT 0,
    "balance" INT NOT NULL DEFAULT 0,
    PRIMARY KEY ("user_id"),
    FOREIGN KEY ("user_id") REFERENCES users(user_id)
)
