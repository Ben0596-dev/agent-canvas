import type { GenerationTask, ModelProfile } from "@/domain/generation/types";

const apiBase = "/backend-api/api/v1";

export async function listModels(): Promise<ModelProfile[]> {
  return (await request<{ items: ModelProfile[] }>("/models")).items;
}

export function createGenerationTask(
  projectId: string,
  input: { modelId: string; prompt: string },
): Promise<GenerationTask> {
  return request(`/projects/${encodeURIComponent(projectId)}/generations`, {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function getGenerationTask(taskId: string): Promise<GenerationTask> {
  return request(`/generation-tasks/${encodeURIComponent(taskId)}`);
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${apiBase}${path}`, {
    ...init,
    headers: { "Content-Type": "application/json", ...init?.headers },
  });
  if (!response.ok) {
    const body = (await response.json().catch(() => ({}))) as { message?: string };
    throw new Error(body.message ?? `Request failed with ${response.status}`);
  }
  return (await response.json()) as T;
}
