CREATE TABLE IF NOT EXISTS tasks (
    id          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    team_id     BIGINT UNSIGNED NOT NULL,
    title       VARCHAR(255)    NOT NULL,
    description TEXT                NULL,
    status      VARCHAR(20)     NOT NULL,
    priority    VARCHAR(20)     NOT NULL,
    assignee_id BIGINT UNSIGNED     NULL,
    created_by  BIGINT UNSIGNED NOT NULL,
    due_date    DATETIME            NULL,
    created_at  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    PRIMARY KEY (id),
    CONSTRAINT fk_tasks_team     FOREIGN KEY (team_id)     REFERENCES teams (id),
    CONSTRAINT fk_tasks_assignee FOREIGN KEY (assignee_id) REFERENCES users (id),
    CONSTRAINT fk_tasks_creator  FOREIGN KEY (created_by)  REFERENCES users (id)
);
