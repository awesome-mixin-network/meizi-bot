CREATE TABLE IF NOT EXISTS subscribers (
  user_id          VARCHAR(36) PRIMARY KEY,
  created_at       TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX subscribers_user_id ON subscribers (user_id);

CREATE TABLE `items` (
    `id`    INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE,
    `item_id`   TEXT NOT NULL UNIQUE,
    `name`  VARCHAR(64),
    `desc`  VARCHAR(1024),
    `ref`   VARCHAR(1024),
    `created_at`    TIMESTAMP NOT NULL,
    `urls`  TEXT,
    `service_name`  VARCHAR(64) NOT NULL
)
CREATE INDEX items_item_id ON items (item_id);

