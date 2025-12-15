CREATE TABLE users (
    "user_id"     BIGINT NOT NULL,
    "name"        TEXT NOT NULL,
    "nick_name"   TEXT,
    "summary"     TEXT,
    "token_usage" BIGINT NOT NULL DEFAULT 0,
    PRIMARY KEY ("user_id")
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
    PRIMARY KEY ("group_id")
);

CREATE INDEX idx_messages_covering ON messages("group_id", "is_cmd", "time" DESC, "user_id", "content", "msg_id");

CREATE TABLE group_rcon_configs (
    "group_id" BIGINT NOT NULL,
    "address"  TEXT NOT NULL,
    "password" TEXT NOT NULL,
    "enabled"  BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY ("group_id")
);
