package model

const (
	NodeTypeText        = "text"
	OperationNodeCreate = "node.create"
	OperationNodeUpdate = "node.update"
	OperationNodeMove   = "node.move"
	OperationNodeDelete = "node.delete"
	OperationEdgeCreate = "edge.create"
	OperationEdgeDelete = "edge.delete"
)

type CanvasSnapshot struct {
	Nodes    []CanvasNode `json:"nodes"`
	Edges    []CanvasEdge `json:"edges"`
	Viewport Viewport     `json:"viewport"`
}

type CanvasNode struct {
	ID       string         `json:"id"`
	Type     string         `json:"type"`
	Title    string         `json:"title"`
	Position CanvasPoint    `json:"position"`
	Size     CanvasSize     `json:"size"`
	Data     CanvasNodeData `json:"data"`
}

type CanvasEdge struct {
	ID     string `json:"id"`
	Source string `json:"source"`
	Target string `json:"target"`
}

type CanvasPoint struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type CanvasSize struct {
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

type Viewport struct {
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
	Zoom float64 `json:"zoom"`
}

type CanvasNodeData struct {
	Text string `json:"text"`
}

type CanvasOperation struct {
	Type     string          `json:"type"`
	Node     *CanvasNode     `json:"node,omitempty"`
	NodeID   string          `json:"nodeId,omitempty"`
	Title    *string         `json:"title,omitempty"`
	Data     *CanvasNodeData `json:"data,omitempty"`
	Position *CanvasPoint    `json:"position,omitempty"`
	Edge     *CanvasEdge     `json:"edge,omitempty"`
	EdgeID   string          `json:"edgeId,omitempty"`
}
