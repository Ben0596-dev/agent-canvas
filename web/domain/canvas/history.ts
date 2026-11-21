import { applyCanvasOperations } from "./reducer";
import type { CanvasOperation, CanvasSnapshot } from "./types";

export type HistoryEntry = {
  forward: CanvasOperation[];
  inverse: CanvasOperation[];
};

export function createHistoryEntry(
  snapshot: CanvasSnapshot,
  operations: CanvasOperation[],
): HistoryEntry {
  return { forward: operations, inverse: invertCanvasOperations(snapshot, operations) };
}

export function invertCanvasOperations(
  snapshot: CanvasSnapshot,
  operations: CanvasOperation[],
): CanvasOperation[] {
  let current = structuredClone(snapshot);
  const inverseBatches: CanvasOperation[][] = [];

  for (const operation of operations) {
    inverseBatches.push(invertOperation(current, operation));
    current = applyCanvasOperations(current, [operation]);
  }

  return inverseBatches.reverse().flat();
}

function invertOperation(snapshot: CanvasSnapshot, operation: CanvasOperation): CanvasOperation[] {
  switch (operation.type) {
    case "node.create":
      return [{ type: "node.delete", nodeId: operation.node.id }];
    case "node.update": {
      const node = requireNode(snapshot, operation.nodeId);
      return [
        {
          type: "node.update",
          nodeId: node.id,
          ...(operation.title !== undefined ? { title: node.title } : {}),
          ...(operation.data !== undefined ? { data: node.data } : {}),
        },
      ];
    }
    case "node.move": {
      const node = requireNode(snapshot, operation.nodeId);
      return [{ type: "node.move", nodeId: node.id, position: node.position }];
    }
    case "node.delete": {
      const node = requireNode(snapshot, operation.nodeId);
      const relatedEdges = snapshot.edges.filter(
        (edge) => edge.source === node.id || edge.target === node.id,
      );
      return [
        { type: "node.create", node },
        ...relatedEdges.map((edge): CanvasOperation => ({ type: "edge.create", edge })),
      ];
    }
    case "edge.create":
      return [{ type: "edge.delete", edgeId: operation.edge.id }];
    case "edge.delete": {
      const edge = snapshot.edges.find((candidate) => candidate.id === operation.edgeId);
      if (!edge) throw new Error(`edge ${operation.edgeId} does not exist`);
      return [{ type: "edge.create", edge }];
    }
  }
}

function requireNode(snapshot: CanvasSnapshot, nodeId: string) {
  const node = snapshot.nodes.find((candidate) => candidate.id === nodeId);
  if (!node) throw new Error(`node ${nodeId} does not exist`);
  return node;
}
