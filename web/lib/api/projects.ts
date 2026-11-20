import type { CanvasOperation, Project, ProjectSummary } from "@/domain/canvas/types";

const apiBase = "/backend-api/api/v1";

type ApiErrorBody = { code?: string; message?: string; currentRevision?: number };

export class ApiError extends Error {
  constructor(
    message: string,
    readonly status: number,
    readonly body: ApiErrorBody,
  ) {
    super(message);
  }
}

export async function listProjects(): Promise<ProjectSummary[]> {
  return (await request<{ items: ProjectSummary[] }>("/projects")).items;
}

export function createProject(title: string): Promise<Project> {
  return request("/projects", { method: "POST", body: JSON.stringify({ title }) });
}

export function getProject(projectId: string): Promise<Project> {
  return request(`/projects/${encodeURIComponent(projectId)}`);
}

export function deleteProject(projectId: string): Promise<void> {
  return request(`/projects/${encodeURIComponent(projectId)}`, { method: "DELETE" });
}

export async function applyOperations(
  projectId: string,
  expectedRevision: number,
  operations: CanvasOperation[],
): Promise<Project> {
  const response = await request<{ revision: number; project: Project }>(
    `/projects/${encodeURIComponent(projectId)}/operations`,
    {
      method: "POST",
      body: JSON.stringify({ requestId: crypto.randomUUID(), expectedRevision, operations }),
    },
  );
  return response.project;
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${apiBase}${path}`, {
    ...init,
    headers: { "Content-Type": "application/json", ...init?.headers },
  });
  if (!response.ok) {
    const body = (await response.json().catch(() => ({}))) as ApiErrorBody;
    throw new ApiError(
      body.message ?? `Request failed with ${response.status}`,
      response.status,
      body,
    );
  }
  if (response.status === 204) return undefined as T;
  return (await response.json()) as T;
}
