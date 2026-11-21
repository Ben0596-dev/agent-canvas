package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"agent-canvas/backend/internal/model"
	"agent-canvas/backend/internal/repository"
)

type OpenAICompatibleProvider struct {
	baseURL string
	apiKey  string
	model   string
	client  *http.Client
}

func NewOpenAICompatibleProvider(baseURL string, apiKey string, model string, client *http.Client) *OpenAICompatibleProvider {
	return &OpenAICompatibleProvider{
		baseURL: strings.TrimRight(baseURL, "/"), apiKey: apiKey, model: model, client: client,
	}
}

func (provider *OpenAICompatibleProvider) Generate(ctx context.Context, request repository.AgentRequest) (repository.AgentResponse, error) {
	payload := map[string]any{
		"model": provider.model,
		"messages": []map[string]string{
			{"role": "system", "content": agentSystemPrompt},
			{"role": "user", "content": fmt.Sprintf(
				"Current canvas JSON:\n%s\n\nSelected node IDs: %v\n\nUser goal:\n%s",
				request.Project.SnapshotJSON, request.SelectedNodeIDs, request.Message,
			)},
		},
		"response_format": map[string]string{"type": "json_object"},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return repository.AgentResponse{}, err
	}
	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, provider.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return repository.AgentResponse{}, err
	}
	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.Header.Set("Authorization", "Bearer "+provider.apiKey)
	response, err := provider.client.Do(httpRequest)
	if err != nil {
		return repository.AgentResponse{}, err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		message, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		return repository.AgentResponse{}, fmt.Errorf("model provider returned %d: %s", response.StatusCode, strings.TrimSpace(string(message)))
	}
	var completion struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(response.Body).Decode(&completion); err != nil {
		return repository.AgentResponse{}, fmt.Errorf("decode model response: %w", err)
	}
	if len(completion.Choices) == 0 {
		return repository.AgentResponse{}, fmt.Errorf("model provider returned no choices")
	}
	content := strings.TrimSpace(completion.Choices[0].Message.Content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	var proposal struct {
		Explanation string                  `json:"explanation"`
		Operations  []model.CanvasOperation `json:"operations"`
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(content)), &proposal); err != nil {
		return repository.AgentResponse{}, fmt.Errorf("decode Agent proposal: %w", err)
	}
	if strings.TrimSpace(proposal.Explanation) == "" {
		return repository.AgentResponse{}, fmt.Errorf("Agent proposal explanation is empty")
	}
	return repository.AgentResponse{Explanation: proposal.Explanation, Operations: proposal.Operations}, nil
}

const agentSystemPrompt = `You operate an infinite canvas by proposing JSON operations. Return one JSON object with exactly:
{"explanation":"short user-facing explanation","operations":[...]}

Supported operation types are node.create, node.update, node.move, node.delete, edge.create, and edge.delete.
Text nodes require id, type="text", title, position{x,y}, size{width,height}, and data{text}.
Edges require id, source, and target referencing existing or newly-created node IDs.
Use unique IDs, positive node sizes, non-empty text, no self-connections, and at most 100 operations.
Return an empty operations array when the user only asks a question. Never include markdown or execute changes directly.`
