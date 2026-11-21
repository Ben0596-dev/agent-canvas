CREATE TABLE IF NOT EXISTS agent_sessions (
    id VARCHAR(64) PRIMARY KEY,
    project_id VARCHAR(36) NOT NULL REFERENCES canvas_projects(id) ON DELETE CASCADE,
    title VARCHAR(200) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_agent_sessions_project_updated
    ON agent_sessions (project_id, updated_at DESC);

CREATE TABLE IF NOT EXISTS agent_runs (
    id VARCHAR(64) PRIMARY KEY,
    project_id VARCHAR(36) NOT NULL REFERENCES canvas_projects(id) ON DELETE CASCADE,
    session_id VARCHAR(64) NOT NULL REFERENCES agent_sessions(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL,
    message TEXT NOT NULL,
    selected_nodes_json JSONB NOT NULL DEFAULT '[]'::jsonb,
    expected_revision BIGINT NOT NULL,
    explanation TEXT NOT NULL DEFAULT '',
    operations_json JSONB NOT NULL DEFAULT '[]'::jsonb,
    error_message TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_agent_runs_session_created
    ON agent_runs (session_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_agent_runs_project_status
    ON agent_runs (project_id, status);
