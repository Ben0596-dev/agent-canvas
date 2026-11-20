import { describe, expect, it } from "vitest";

import { applyCanvasOperations } from "./reducer";
import type { CanvasSnapshot } from "./types";

describe("applyCanvasOperations", () => {
  it("applies a batch without mutating the original snapshot", () => {
    const original: CanvasSnapshot = {
      nodes: [],
      edges: [],
      viewport: { x: 0, y: 0, zoom: 1 },
    };

    const result = applyCanvasOperations(original, [
      {
        type: "node.create",
        node: {
          id: "node-1",
          type: "text",
          title: "First step",
          position: { x: 100, y: 200 },
          size: { width: 240, height: 120 },
          data: { text: "Define the goal" },
        },
      },
      { type: "node.move", nodeId: "node-1", position: { x: 320, y: 160 } },
    ]);

    expect(original.nodes).toHaveLength(0);
    expect(result.nodes[0]?.position).toEqual({ x: 320, y: 160 });
  });
});
