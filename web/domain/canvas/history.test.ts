import { expect, it } from "vitest";

import { applyCanvasOperations } from "./reducer";
import { invertCanvasOperations } from "./history";
import type { CanvasOperation, CanvasSnapshot } from "./types";

it("builds an inverse batch that restores nodes and related edges", () => {
  const original: CanvasSnapshot = {
    nodes: [
      {
        id: "node-1",
        type: "text",
        title: "Draft",
        position: { x: 10, y: 20 },
        size: { width: 240, height: 120 },
        data: { text: "Draft" },
      },
      {
        id: "node-2",
        type: "text",
        title: "Review",
        position: { x: 300, y: 20 },
        size: { width: 240, height: 120 },
        data: { text: "Review" },
      },
    ],
    edges: [{ id: "edge-1", source: "node-1", target: "node-2" }],
    viewport: { x: 0, y: 0, zoom: 1 },
  };
  const forward: CanvasOperation[] = [
    { type: "node.move", nodeId: "node-1", position: { x: 80, y: 90 } },
    { type: "node.delete", nodeId: "node-2" },
  ];

  const changed = applyCanvasOperations(original, forward);
  const restored = applyCanvasOperations(changed, invertCanvasOperations(original, forward));

  expect(restored).toEqual(original);
});
