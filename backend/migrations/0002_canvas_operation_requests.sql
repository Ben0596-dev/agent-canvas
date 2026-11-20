CREATE TABLE IF NOT EXISTS canvas_operation_requests (
    project_id VARCHAR(36) NOT NULL REFERENCES canvas_projects(id) ON DELETE CASCADE,
    request_id VARCHAR(64) NOT NULL,
    result_revision BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (project_id, request_id)
);
