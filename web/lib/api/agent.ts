import type { AgentRun, AgentSession } from "@/domain/agent/types";

const apiBase = "/backend-api/api/v1";

export async function createAgentRun(
  projectId: string,
  message: string,
  sessionId?: string,
): Promise<AgentRun> {
  return request(`/projects/${encodeURIComponent(projectId)}/agent/runs`, {
    method: "POST",
    body: JSON.stringify({ message, sessionId }),
  });
}

export async function listAgentSessions(projectId: string): Promise<AgentSession[]> {
  return (
    await request<{ items: AgentSession[] }>(
      `/projects/${encodeURIComponent(projectId)}/agent/sessions`,
    )
  ).items;
}

export async function listAgentRuns(projectId: string, sessionId: string): Promise<AgentRun[]> {
  return (
    await request<{ items: AgentRun[] }>(
      `/projects/${encodeURIComponent(projectId)}/agent/sessions/${encodeURIComponent(sessionId)}/runs`,
    )
  ).items;
}

export function cancelAgentRun(runId: string): Promise<AgentRun> {
  return request(`/agent/runs/${encodeURIComponent(runId)}/cancel`, { method: "POST" });
}

export function agentRunEventsURL(runId: string): string {
  return `${apiBase}/agent/runs/${encodeURIComponent(runId)}/events`;
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
