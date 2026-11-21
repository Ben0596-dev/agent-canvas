package agent

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"agent-canvas/backend/internal/model"
	"agent-canvas/backend/internal/repository"
)

func TestOpenAICompatibleProviderParsesStructuredProposal(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.Header.Get("Authorization") != "Bearer secret" {
			t.Fatalf("missing authorization header")
		}
		var body struct {
			Model string `json:"model"`
		}
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if body.Model != "agent-model" {
			t.Fatalf("unexpected model %q", body.Model)
		}
		response.Header().Set("Content-Type", "application/json")
		_, _ = response.Write([]byte(`{
			"choices":[{"message":{"content":"{\"explanation\":\"创建目标节点\",\"operations\":[{\"type\":\"node.create\",\"node\":{\"id\":\"node-1\",\"type\":\"text\",\"title\":\"目标\",\"position\":{\"x\":120,\"y\":120},\"size\":{\"width\":240,\"height\":140},\"data\":{\"text\":\"发布产品\"}}}]}"}}]
		}`))
	}))
	defer server.Close()

	provider := NewOpenAICompatibleProvider(server.URL, "secret", "agent-model", server.Client())
	result, err := provider.Generate(context.Background(), repository.AgentRequest{
		Project: model.Project{SnapshotJSON: []byte(`{"nodes":[],"edges":[],"viewport":{"x":0,"y":0,"zoom":1}}`)},
		Message: "制定发布目标",
	})
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if result.Explanation != "创建目标节点" || len(result.Operations) != 1 {
		t.Fatalf("unexpected proposal: %+v", result)
	}
}
