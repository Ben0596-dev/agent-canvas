"use client";

import { useEffect, useMemo, useState } from "react";

import type { GenerationTask, ModelProfile } from "@/domain/generation/types";
import { createGenerationTask, getGenerationTask, listModels } from "@/lib/api/generation";

export function GenerationPanel({
  projectId,
  onGenerated,
}: {
  projectId: string;
  onGenerated: (task: GenerationTask) => Promise<boolean>;
}) {
  const [models, setModels] = useState<ModelProfile[]>([]);
  const [kind, setKind] = useState<"image" | "video">("image");
  const [modelId, setModelId] = useState("");
  const [prompt, setPrompt] = useState("");
  const [status, setStatus] = useState("");
  const available = useMemo(
    () => models.filter((model) => model.capability === kind),
    [kind, models],
  );
  const selectedModelId = available.some((model) => model.id === modelId)
    ? modelId
    : (available[0]?.id ?? "");

  useEffect(() => {
    void listModels()
      .then(setModels)
      .catch((error: unknown) =>
        setStatus(error instanceof Error ? error.message : "模型载入失败"),
      );
  }, []);

  async function generate() {
    if (!selectedModelId || !prompt.trim()) return;
    try {
      let task = await createGenerationTask(projectId, { modelId: selectedModelId, prompt });
      setStatus("生成任务已提交…");
      const deadline = Date.now() + 10 * 60 * 1000;
      while (Date.now() < deadline && (task.status === "queued" || task.status === "running")) {
        await new Promise((resolve) => window.setTimeout(resolve, 1500));
        task = await getGenerationTask(task.id);
        setStatus(task.status === "running" ? "生成中…" : task.status);
      }
      if (task.status === "failed") throw new Error(task.error || "生成失败");
      if (task.status !== "succeeded") throw new Error("生成任务等待超时");
      if (await onGenerated(task)) {
        setStatus("结果已加入画布");
        setPrompt("");
      }
    } catch (error) {
      setStatus(error instanceof Error ? error.message : "生成失败");
    }
  }

  return (
    <section className="border-t border-slate-200 p-4">
      <div className="mb-3 flex items-center justify-between">
        <h3 className="text-sm font-semibold text-slate-900">媒体生成</h3>
        <select
          className="rounded-md border border-slate-200 px-2 py-1 text-xs"
          value={kind}
          onChange={(event) => setKind(event.target.value as "image" | "video")}
        >
          <option value="image">图片</option>
          <option value="video">视频</option>
        </select>
      </div>
      <select
        className="mb-2 w-full rounded-lg border border-slate-200 px-3 py-2 text-xs"
        value={selectedModelId}
        onChange={(event) => setModelId(event.target.value)}
      >
        {available.map((model) => (
          <option key={model.id} value={model.id}>
            {model.upstreamModel}
          </option>
        ))}
      </select>
      <textarea
        className="h-20 w-full resize-none rounded-lg border border-slate-200 p-3 text-sm outline-none focus:border-indigo-500"
        placeholder={kind === "image" ? "描述要生成的图片" : "描述要生成的视频"}
        value={prompt}
        onChange={(event) => setPrompt(event.target.value)}
      />
      <div className="mt-2 flex items-center justify-between gap-2">
        <span className="truncate text-xs text-slate-500">{status}</span>
        <button
          className="rounded-lg bg-indigo-600 px-3 py-2 text-xs font-medium text-white"
          onClick={() => void generate()}
        >
          开始生成
        </button>
      </div>
    </section>
  );
}
