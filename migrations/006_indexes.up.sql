CREATE INDEX idx_tasks_team_created_at ON tasks (team_id, created_at);
CREATE INDEX idx_task_comments_task_id ON task_comments (task_id);
CREATE INDEX idx_task_history_task_id ON task_history (task_id);
