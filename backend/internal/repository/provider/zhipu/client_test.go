package zhipu

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"agent-canvas/backend/internal/repository"
)

func TestGenerateImageUsesCogViewAndReturnsTemporaryURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/images/generations" {
			t.Fatalf("unexpected path %s", request.URL.Path)
		}
		if request.Header.Get("Authorization") != "Bearer secret" {
			t.Fatal("missing authorization")
		}
		var body map[string]any
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if body["model"] != "cogview-3-flash" || body["size"] != "1024x1024" {
			t.Fatalf("unexpected request: %+v", body)
		}
		response.Header().Set("Content-Type", "application/json")
		_, _ = response.Write([]byte(`{"created":1,"data":[{"url":"https://cdn.example/image.png"}]}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "secret", server.Client())
	result, err := client.GenerateImage(context.Background(), repository.ImageGenerationRequest{
		Model: "cogview-3-flash", Prompt: "未来城市", Size: "1024x1024",
	})
	if err != nil {
		t.Fatalf("generate image: %v", err)
	}
	if result.URL != "https://cdn.example/image.png" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestSubmitAndPollCogVideoTask(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("Content-Type", "application/json")
		switch {
		case request.Method == http.MethodPost && request.URL.Path == "/videos/generations":
			var body map[string]any
			if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
				t.Fatalf("decode request: %v", err)
			}
			if body["model"] != "cogvideox-flash" || body["request_id"] != "task-1" {
				t.Fatalf("unexpected request: %+v", body)
			}
			_, _ = response.Write([]byte(`{"id":"provider-task-1","task_status":"PROCESSING"}`))
		case request.Method == http.MethodGet && request.URL.Path == "/async-result/provider-task-1":
			_, _ = response.Write([]byte(`{"task_status":"SUCCESS","video_result":[{"url":"https://cdn.example/video.mp4","cover_image_url":"https://cdn.example/cover.jpg"}]}`))
		default:
			http.NotFound(response, request)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, "secret", server.Client())
	submitted, err := client.SubmitVideo(context.Background(), repository.VideoGenerationRequest{
		RequestID: "task-1", Model: "cogvideox-flash", Prompt: "猫在奔跑", Size: "1920x1080", FPS: 30,
	})
	if err != nil {
		t.Fatalf("submit video: %v", err)
	}
	if submitted.ProviderTaskID != "provider-task-1" || submitted.Status != repository.GenerationRunning {
		t.Fatalf("unexpected submitted task: %+v", submitted)
	}
	result, err := client.GetVideoResult(context.Background(), submitted.ProviderTaskID)
	if err != nil {
		t.Fatalf("poll video: %v", err)
	}
	if result.Status != repository.GenerationSucceeded || result.URL != "https://cdn.example/video.mp4" {
		t.Fatalf("unexpected result: %+v", result)
	}
}
