CREATE TABLE IF NOT EXISTS task_history (
    id        BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    task_id   BIGINT UNSIGNED NOT NULL,
    user_id   BIGINT UNSIGNED NOT NULL,
    field     VARCHAR(50)     NOT NULL,
    old_value TEXT                NULL,
    new_value TEXT                NULL,
    created_at DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (id),
    CONSTRAINT fk_task_history_task FOREIGN KEY (task_id) REFERENCES tasks (id),
    CONSTRAINT fk_task_history_user FOREIGN KEY (user_id) REFERENCES users (id)
);

CREATE TABLE IF NOT EXISTS task_comments (
    id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    task_id    BIGINT UNSIGNED NOT NULL,
    user_id    BIGINT UNSIGNED NOT NULL,
    body       TEXT            NOT NULL,
    created_at DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (id),
    CONSTRAINT fk_task_comments_task FOREIGN KEY (task_id) REFERENCES tasks (id),
    CONSTRAINT fk_task_comments_user FOREIGN KEY (user_id) REFERENCES users (id)
);
