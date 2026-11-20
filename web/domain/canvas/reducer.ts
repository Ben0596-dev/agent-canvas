import type { CanvasOperation, CanvasSnapshot } from "./types";

export function applyCanvasOperations(
  snapshot: CanvasSnapshot,
  operations: CanvasOperation[],
): CanvasSnapshot {
  const next = structuredClone(snapshot);

  for (const operation of operations) {
    switch (operation.type) {
      case "node.create":
        next.nodes.push(operation.node);
        break;
      case "node.update": {
        const node = requireNode(next, operation.nodeId);
        if (operation.title !== undefined) node.title = operation.title;
        if (operation.data !== undefined) node.data = operation.data;
        break;
      }
      case "node.move":
        requireNode(next, operation.nodeId).position = operation.position;
        break;
      case "node.delete":
        next.nodes = next.nodes.filter((node) => node.id !== operation.nodeId);
        next.edges = next.edges.filter(
          (edge) => edge.source !== operation.nodeId && edge.target !== operation.nodeId,
        );
        break;
      case "edge.create":
        next.edges.push(operation.edge);
        break;
      case "edge.delete":
        next.edges = next.edges.filter((edge) => edge.id !== operation.edgeId);
        break;
    }
  }

  return next;
}

function requireNode(snapshot: CanvasSnapshot, nodeId: string) {
  const node = snapshot.nodes.find((candidate) => candidate.id === nodeId);
  if (!node) throw new Error(`node ${nodeId} does not exist`);
  return node;
}
