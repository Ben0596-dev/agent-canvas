import type { CanvasOperation } from "../canvas/types";

export type AgentProposal = {
  explanation: string;
  expectedRevision: number;
  operations: CanvasOperation[];
};

export type AgentRun = {
  id: string;
  projectId: string;
  sessionId: string;
  status: "queued" | "running" | "succeeded" | "failed" | "cancelled";
  message: string;
  proposal?: AgentProposal;
  error?: string;
  createdAt: string;
  updatedAt: string;
};

export type AgentSession = {
  id: string;
  projectId: string;
  title: string;
  createdAt: string;
  updatedAt: string;
};
