package agent

import (
	"context"
	"crypto/rand"
	"encoding/hex"

	"agent-canvas/backend/internal/model"
	"agent-canvas/backend/internal/repository"
)

type LocalProvider struct{}

func (LocalProvider) Generate(ctx context.Context, request repository.AgentRequest) (repository.AgentResponse, error) {
	if err := ctx.Err(); err != nil {
		return repository.AgentResponse{}, err
	}
	return repository.AgentResponse{
		Explanation: "已生成一个文本节点。请预览并确认后写入画布。",
		Operations: []model.CanvasOperation{{
			Type: model.OperationNodeCreate,
			Node: &model.CanvasNode{
				ID:       newID(),
				Type:     model.NodeTypeText,
				Title:    "Agent 提议",
				Position: model.CanvasPoint{X: 120 + float64(len(request.Project.SnapshotJSON)%8)*24, Y: 120},
				Size:     model.CanvasSize{Width: 280, Height: 160},
				Data:     model.CanvasNodeData{Text: request.Message},
			},
		}},
	}, nil
}

func newID() string {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		panic(err)
	}
	return hex.EncodeToString(raw[:])
}
