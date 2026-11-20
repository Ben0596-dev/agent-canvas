"use client";

import "@xyflow/react/dist/style.css";

import {
  Background,
  BackgroundVariant,
  Controls,
  MiniMap,
  ReactFlow,
  addEdge,
  useEdgesState,
  useNodesState,
  type Connection,
  type Edge,
  type Node,
  type NodeTypes,
} from "@xyflow/react";
import Link from "next/link";
import { useCallback, useEffect, useMemo, useState } from "react";

import type { CanvasOperation, Project } from "@/domain/canvas/types";
import { applyOperations, getProject } from "@/lib/api/projects";

import { TextCanvasNode, type TextCanvasNodeData } from "./text-canvas-node";

type TextFlowNode = Node<TextCanvasNodeData, "text">;
const nodeTypes: NodeTypes = { text: TextCanvasNode };

export function CanvasWorkspace({ projectId }: { projectId: string }) {
  const [project, setProject] = useState<Project | null>(null);
  const [nodes, setNodes, onNodesChange] = useNodesState<TextFlowNode>([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState<Edge>([]);
  const [status, setStatus] = useState("正在载入画布…");

  const commit = useCallback(
    async (operations: CanvasOperation[]) => {
      if (!project || operations.length === 0) return;
      try {
        const updated = await applyOperations(project.id, project.revision, operations);
        setProject(updated);
        setStatus("已保存");
      } catch (error) {
        setStatus(error instanceof Error ? error.message : "保存失败");
        const latest = await getProject(project.id).catch(() => null);
        if (latest) setProject(latest);
      }
    },
    [project],
  );

  useEffect(() => {
    void getProject(projectId)
      .then((loaded) => {
        setProject(loaded);
        setStatus("");
      })
      .catch((error: unknown) =>
        setStatus(error instanceof Error ? error.message : "画布载入失败"),
      );
  }, [projectId]);

  useEffect(() => {
    if (!project) return;
    setNodes(
      project.snapshot.nodes.map((node) => ({
        id: node.id,
        type: "text" as const,
        position: node.position,
        style: { width: node.size.width, height: node.size.height },
        data: {
          title: node.title,
          text: node.data.text,
          onCommit: (text: string) =>
            void commit([{ type: "node.update", nodeId: node.id, data: { text } }]),
        },
      })),
    );
    setEdges(project.snapshot.edges.map((edge) => ({ ...edge })));
  }, [commit, project, setEdges, setNodes]);

  const onConnect = useCallback(
    (connection: Connection) => {
      if (!connection.source || !connection.target) return;
      const edge = {
        id: crypto.randomUUID(),
        source: connection.source,
        target: connection.target,
      };
      setEdges((current) => addEdge(edge, current));
      void commit([{ type: "edge.create", edge }]);
    },
    [commit, setEdges],
  );
  const initialViewport = useMemo(
    () => project?.snapshot.viewport ?? { x: 0, y: 0, zoom: 1 },
    [project],
  );

  function addTextNode() {
    void commit([
      {
        type: "node.create",
        node: {
          id: crypto.randomUUID(),
          type: "text",
          title: "新文本",
          position: { x: 120 + nodes.length * 24, y: 120 + nodes.length * 24 },
          size: { width: 240, height: 140 },
          data: { text: "在这里输入内容" },
        },
      },
    ]);
  }

  return (
    <main className="flex h-screen min-h-[640px] flex-col bg-slate-100">
      <header className="flex h-16 items-center justify-between border-b border-slate-200 bg-white px-5">
        <div className="flex items-center gap-4">
          <Link href="/" className="text-sm text-slate-500">
            项目
          </Link>
          <h1 className="font-semibold text-slate-950">{project?.title ?? "画布"}</h1>
          {project ? <span className="text-xs text-slate-400">r{project.revision}</span> : null}
        </div>
        <div className="flex items-center gap-4">
          <span className="text-xs text-slate-500">{status}</span>
          <button
            className="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white"
            onClick={addTextNode}
          >
            添加文本
          </button>
        </div>
      </header>
      <section className="min-h-0 flex-1">
        <ReactFlow<TextFlowNode, Edge>
          nodes={nodes}
          edges={edges}
          nodeTypes={nodeTypes}
          defaultViewport={initialViewport}
          onNodesChange={onNodesChange}
          onEdgesChange={onEdgesChange}
          onConnect={onConnect}
          onNodeDragStop={(_, node) =>
            void commit([{ type: "node.move", nodeId: node.id, position: node.position }])
          }
          onNodesDelete={(deleted) =>
            void commit(deleted.map((node) => ({ type: "node.delete", nodeId: node.id })))
          }
          onEdgesDelete={(deleted) =>
            void commit(deleted.map((edge) => ({ type: "edge.delete", edgeId: edge.id })))
          }
          fitView={nodes.length > 0}
        >
          <Background variant={BackgroundVariant.Dots} gap={20} size={1.2} />
          <Controls />
          <MiniMap pannable zoomable />
        </ReactFlow>
      </section>
    </main>
  );
}
