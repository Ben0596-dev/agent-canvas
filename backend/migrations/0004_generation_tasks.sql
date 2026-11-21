CREATE TABLE IF NOT EXISTS generation_tasks (
    id VARCHAR(64) PRIMARY KEY,
    project_id VARCHAR(36) NOT NULL REFERENCES canvas_projects(id) ON DELETE CASCADE,
    kind VARCHAR(20) NOT NULL,
    provider VARCHAR(50) NOT NULL,
    model_id VARCHAR(100) NOT NULL,
    upstream_model VARCHAR(100) NOT NULL,
    status VARCHAR(20) NOT NULL,
    prompt TEXT NOT NULL,
    parameters_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    provider_task_id VARCHAR(200) NOT NULL DEFAULT '',
    result_url TEXT NOT NULL DEFAULT '',
    cover_url TEXT NOT NULL DEFAULT '',
    error_message TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_generation_tasks_project_created
    ON generation_tasks (project_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_generation_tasks_status
    ON generation_tasks (status);
