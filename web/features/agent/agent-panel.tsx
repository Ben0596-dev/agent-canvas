"use client";

import { useEffect, useRef, useState } from "react";

import type { AgentProposal, AgentRun } from "@/domain/agent/types";
import {
  agentRunEventsURL,
  cancelAgentRun,
  createAgentRun,
  listAgentRuns,
  listAgentSessions,
} from "@/lib/api/agent";

type AgentPanelProps = {
  projectId: string;
  proposal: AgentProposal | null;
  onProposalChange: (proposal: AgentProposal | null) => void;
  onConfirm: (proposal: AgentProposal) => Promise<boolean>;
};

export function AgentPanel({ projectId, proposal, onProposalChange, onConfirm }: AgentPanelProps) {
  const [message, setMessage] = useState("");
  const [run, setRun] = useState<AgentRun | null>(null);
  const [history, setHistory] = useState<AgentRun[]>([]);
  const [sessionId, setSessionId] = useState<string>();
  const [status, setStatus] = useState("");
  const events = useRef<EventSource | null>(null);

  useEffect(() => () => events.current?.close(), []);

  useEffect(() => {
    void listAgentSessions(projectId)
      .then(async (sessions) => {
        const latest = sessions[0];
        if (!latest) return;
        setSessionId(latest.id);
        setHistory(await listAgentRuns(projectId, latest.id));
      })
      .catch((error: unknown) =>
        setStatus(error instanceof Error ? error.message : "会话恢复失败"),
      );
  }, [projectId]);

  async function submit() {
    if (!message.trim()) return;
    events.current?.close();
    onProposalChange(null);
    setStatus("Agent 正在分析画布…");
    try {
      const created = await createAgentRun(projectId, message, sessionId);
      setRun(created);
      setSessionId(created.sessionId);
      setHistory((current) => [...current, created]);
      const source = new EventSource(agentRunEventsURL(created.id));
      events.current = source;
      source.addEventListener("operation.proposed", (event) => {
        const next = JSON.parse((event as MessageEvent<string>).data) as AgentProposal;
        onProposalChange(next);
        setStatus("变更已生成，等待确认");
      });
      source.addEventListener("run.completed", (event) => {
        const completed = JSON.parse((event as MessageEvent<string>).data) as AgentRun;
        setRun(completed);
        setHistory((current) =>
          current.map((item) => (item.id === completed.id ? completed : item)),
        );
        source.close();
      });
      source.addEventListener("run.failed", (event) => {
        const failed = JSON.parse((event as MessageEvent<string>).data) as AgentRun;
        setRun(failed);
        setHistory((current) => current.map((item) => (item.id === failed.id ? failed : item)));
        setStatus(failed.error || "Agent 执行失败");
        source.close();
      });
      source.addEventListener("run.cancelled", () => {
        setStatus("已取消");
        source.close();
      });
      source.onerror = () => {
        if (!proposal) setStatus("连接中断，可重新发起");
        source.close();
      };
    } catch (error) {
      setStatus(error instanceof Error ? error.message : "Agent 启动失败");
    }
  }

  async function cancel() {
    if (!run) return;
    const cancelled = await cancelAgentRun(run.id);
    setRun(cancelled);
    setHistory((current) => current.map((item) => (item.id === cancelled.id ? cancelled : item)));
    events.current?.close();
    setStatus("已取消");
  }

  async function confirm() {
    if (!proposal) return;
    if (await onConfirm(proposal)) {
      onProposalChange(null);
      setStatus("变更已提交");
    }
  }

  return (
    <section className="flex min-h-0 flex-1 flex-col bg-white">
      <header className="border-b border-slate-200 px-5 py-4">
        <h2 className="font-semibold text-slate-950">Agent</h2>
        <p className="mt-1 text-xs text-slate-500">Agent 只会提出变更，确认后才写入画布。</p>
      </header>
      <div className="min-h-0 flex-1 overflow-y-auto p-5">
        <div className="mb-5 space-y-3">
          {history.map((item) => (
            <article key={item.id} className="rounded-xl border border-slate-200 p-3 text-sm">
              <p className="text-slate-800">{item.message}</p>
              {item.proposal ? (
                <p className="mt-2 text-slate-500">{item.proposal.explanation}</p>
              ) : item.error ? (
                <p className="mt-2 text-rose-600">{item.error}</p>
              ) : (
                <p className="mt-2 text-xs text-slate-400">{item.status}</p>
              )}
            </article>
          ))}
        </div>
        {proposal ? (
          <section className="rounded-xl border border-indigo-200 bg-indigo-50 p-4">
            <p className="text-sm leading-6 text-slate-700">{proposal.explanation}</p>
            <p className="mt-3 text-xs text-indigo-700">
              {proposal.operations.length} 项变更 · 基于 revision {proposal.expectedRevision}
            </p>
            <div className="mt-4 flex gap-2">
              <button
                className="rounded-lg bg-indigo-600 px-3 py-2 text-sm text-white"
                onClick={() => void confirm()}
              >
                确认应用
              </button>
              <button
                className="rounded-lg border border-slate-300 px-3 py-2 text-sm"
                onClick={() => onProposalChange(null)}
              >
                放弃
              </button>
            </div>
          </section>
        ) : (
          <p className="text-sm text-slate-400">描述你希望 Agent 在画布中完成的工作。</p>
        )}
        {status ? <p className="mt-4 text-xs text-slate-500">{status}</p> : null}
      </div>
      <footer className="border-t border-slate-200 p-4">
        <textarea
          className="h-24 w-full resize-none rounded-xl border border-slate-200 p-3 text-sm outline-none focus:border-indigo-500"
          placeholder="例如：为产品发布创建三个步骤节点并连接起来"
          value={message}
          onChange={(event) => setMessage(event.target.value)}
        />
        <div className="mt-3 flex justify-end gap-2">
          {run && (run.status === "queued" || run.status === "running") ? (
            <button
              className="rounded-lg border border-slate-300 px-3 py-2 text-sm"
              onClick={() => void cancel()}
            >
              取消
            </button>
          ) : null}
          <button
            className="rounded-lg bg-slate-950 px-4 py-2 text-sm font-medium text-white"
            onClick={() => void submit()}
          >
            发送
          </button>
        </div>
      </footer>
    </section>
  );
}
