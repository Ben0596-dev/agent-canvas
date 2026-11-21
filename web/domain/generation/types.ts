export type ModelProfile = {
  id: string;
  provider: string;
  upstreamModel: string;
  capability: "agent" | "image" | "video";
  defaults: Record<string, unknown>;
};

export type GenerationTask = {
  id: string;
  projectId: string;
  kind: "image" | "video";
  modelId: string;
  status: "queued" | "running" | "succeeded" | "failed";
  prompt: string;
  resultUrl?: string;
  coverUrl?: string;
  error?: string;
  createdAt: string;
  updatedAt: string;
};
