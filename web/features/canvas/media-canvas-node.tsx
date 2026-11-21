"use client";

import { Handle, Position, type Node, type NodeProps } from "@xyflow/react";

export type MediaCanvasNodeData = {
  kind: "image" | "video";
  title: string;
  url: string;
  coverUrl?: string;
  prompt?: string;
  preview?: boolean;
};

type MediaFlowNode = Node<MediaCanvasNodeData, "image" | "video">;

export function MediaCanvasNode({ data, selected }: NodeProps<MediaFlowNode>) {
  return (
    <article
      className={`h-full overflow-hidden rounded-xl border bg-white shadow-sm ${data.preview ? "border-dashed border-indigo-400" : selected ? "border-indigo-500" : "border-slate-200"}`}
    >
      <Handle type="target" position={Position.Left} />
      {data.kind === "image" ? (
        // Generated media is provided by a configured backend provider.
        // eslint-disable-next-line @next/next/no-img-element
        <img
          className="h-[calc(100%-48px)] w-full object-cover"
          src={data.url}
          alt={data.prompt || data.title}
        />
      ) : (
        <video
          className="h-[calc(100%-48px)] w-full object-cover"
          src={data.url}
          poster={data.coverUrl}
          controls
        />
      )}
      <div className="truncate px-3 py-3 text-xs font-medium text-slate-700">{data.title}</div>
      <Handle type="source" position={Position.Right} />
    </article>
  );
}
