CREATE TABLE IF NOT EXISTS task_comments (
    id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    task_id    BIGINT UNSIGNED NOT NULL,
    user_id    BIGINT UNSIGNED NOT NULL,
    content    TEXT            NOT NULL,
    created_at DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (id),
    CONSTRAINT fk_task_comments_task FOREIGN KEY (task_id) REFERENCES tasks (id),
    CONSTRAINT fk_task_comments_user FOREIGN KEY (user_id) REFERENCES users (id)
);
