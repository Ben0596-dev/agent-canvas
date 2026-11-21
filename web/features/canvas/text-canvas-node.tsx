"use client";

import { Handle, Position, type Node, type NodeProps } from "@xyflow/react";

export type TextCanvasNodeData = {
  title: string;
  text: string;
  preview?: boolean;
  onCommit: (text: string) => void;
};

type TextFlowNode = Node<TextCanvasNodeData, "text">;

export function TextCanvasNode({ data, selected }: NodeProps<TextFlowNode>) {
  return (
    <article
      className={`h-full rounded-xl border bg-white p-3 shadow-sm ${data.preview ? "border-dashed border-indigo-400 bg-indigo-50/70" : selected ? "border-indigo-500" : "border-slate-200"}`}
    >
      <Handle type="target" position={Position.Left} />
      <p className="mb-2 text-xs font-semibold uppercase tracking-wide text-slate-400">
        {data.title}
      </p>
      <textarea
        key={data.text}
        className="nodrag nowheel h-[calc(100%-24px)] w-full resize-none bg-transparent text-sm leading-6 text-slate-800 outline-none"
        defaultValue={data.text}
        readOnly={data.preview}
        onBlur={(event) => {
          const text = event.currentTarget.value;
          if (text.trim() && text !== data.text) data.onCommit(text);
        }}
      />
      <Handle type="source" position={Position.Right} />
    </article>
  );
}
