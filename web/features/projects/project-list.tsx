"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";

import type { ProjectSummary } from "@/domain/canvas/types";
import { createProject, deleteProject, listProjects } from "@/lib/api/projects";

export function ProjectList() {
  const router = useRouter();
  const [projects, setProjects] = useState<ProjectSummary[]>([]);
  const [title, setTitle] = useState("");
  const [status, setStatus] = useState("正在载入项目…");

  useEffect(() => {
    void listProjects()
      .then((items) => {
        setProjects(items);
        setStatus("");
      })
      .catch((error: unknown) =>
        setStatus(error instanceof Error ? error.message : "项目载入失败"),
      );
  }, []);

  async function create() {
    if (!title.trim()) return;
    try {
      const project = await createProject(title);
      router.push(`/projects/${project.id}`);
    } catch (error) {
      setStatus(error instanceof Error ? error.message : "项目创建失败");
    }
  }

  async function remove(projectId: string) {
    try {
      await deleteProject(projectId);
      setProjects((current) => current.filter((project) => project.id !== projectId));
    } catch (error) {
      setStatus(error instanceof Error ? error.message : "项目删除失败");
    }
  }

  return (
    <main className="mx-auto min-h-screen w-full max-w-6xl px-6 py-12">
      <header className="mb-10 flex flex-col gap-3">
        <span className="text-sm font-semibold tracking-[0.2em] text-indigo-600">AGENT CANVAS</span>
        <h1 className="text-4xl font-semibold tracking-tight text-slate-950">
          把想法铺成一张可执行的画布
        </h1>
        <p className="max-w-2xl text-slate-600">
          创建项目，组织文本节点，并让 Agent 在确认后生成结构化画布变更。
        </p>
      </header>
      <section className="mb-10 flex gap-3 rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
        <input
          className="min-w-0 flex-1 rounded-xl border border-slate-200 px-4 py-3 outline-none focus:border-indigo-500"
          placeholder="项目名称"
          value={title}
          onChange={(event) => setTitle(event.target.value)}
          onKeyDown={(event) => event.key === "Enter" && void create()}
        />
        <button
          className="rounded-xl bg-slate-950 px-5 py-3 font-medium text-white"
          onClick={() => void create()}
        >
          新建项目
        </button>
      </section>
      {status ? <p className="mb-5 text-sm text-slate-500">{status}</p> : null}
      <section className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {projects.map((project) => (
          <article
            key={project.id}
            className="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm"
          >
            <Link className="block" href={`/projects/${project.id}`}>
              <h2 className="mb-2 text-lg font-semibold text-slate-950">{project.title}</h2>
              <p className="text-sm text-slate-500">Revision {project.revision}</p>
            </Link>
            <button className="mt-6 text-sm text-rose-600" onClick={() => void remove(project.id)}>
              删除
            </button>
          </article>
        ))}
      </section>
    </main>
  );
}
