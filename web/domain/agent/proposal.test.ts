import { expect, it } from "vitest";

import { previewProposal } from "./proposal";

it("previews an Agent proposal without changing the saved snapshot", () => {
  const snapshot = { nodes: [], edges: [], viewport: { x: 0, y: 0, zoom: 1 } };
  const preview = previewProposal(snapshot, {
    explanation: "创建目标节点",
    expectedRevision: 1,
    operations: [
      {
        type: "node.create",
        node: {
          id: "agent-node-1",
          type: "text",
          title: "目标",
          position: { x: 120, y: 120 },
          size: { width: 240, height: 140 },
          data: { text: "制定发布目标" },
        },
      },
    ],
  });

  expect(snapshot.nodes).toHaveLength(0);
  expect(preview.nodes).toHaveLength(1);
});
