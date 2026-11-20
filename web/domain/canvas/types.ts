export type CanvasPoint = { x: number; y: number };
export type CanvasSize = { width: number; height: number };
export type CanvasViewport = CanvasPoint & { zoom: number };

export type CanvasNode = {
  id: string;
  type: "text";
  title: string;
  position: CanvasPoint;
  size: CanvasSize;
  data: { text: string };
};

export type CanvasEdge = {
  id: string;
  source: string;
  target: string;
};

export type CanvasSnapshot = {
  nodes: CanvasNode[];
  edges: CanvasEdge[];
  viewport: CanvasViewport;
};

export type CanvasOperation =
  | { type: "node.create"; node: CanvasNode }
  | { type: "node.update"; nodeId: string; title?: string; data?: CanvasNode["data"] }
  | { type: "node.move"; nodeId: string; position: CanvasPoint }
  | { type: "node.delete"; nodeId: string }
  | { type: "edge.create"; edge: CanvasEdge }
  | { type: "edge.delete"; edgeId: string };

export type Project = {
  id: string;
  title: string;
  revision: number;
  snapshot: CanvasSnapshot;
  createdAt: string;
  updatedAt: string;
};

export type ProjectSummary = Omit<Project, "snapshot">;
