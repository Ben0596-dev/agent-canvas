package zhipu

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"agent-canvas/backend/internal/repository"
)

type Client struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

func NewClient(baseURL string, apiKey string, httpClient *http.Client) *Client {
	return &Client{baseURL: strings.TrimRight(baseURL, "/"), apiKey: apiKey, http: httpClient}
}

func (client *Client) GenerateImage(ctx context.Context, request repository.ImageGenerationRequest) (repository.ImageGenerationResult, error) {
	payload := map[string]any{
		"model": request.Model, "prompt": request.Prompt, "size": request.Size,
		"quality": "standard", "watermark_enabled": true,
	}
	var response struct {
		Data []struct {
			URL string `json:"url"`
		} `json:"data"`
	}
	if err := client.doJSON(ctx, http.MethodPost, "/images/generations", payload, &response); err != nil {
		return repository.ImageGenerationResult{}, err
	}
	if len(response.Data) == 0 || response.Data[0].URL == "" {
		return repository.ImageGenerationResult{}, fmt.Errorf("zhipu image response contains no URL")
	}
	return repository.ImageGenerationResult{URL: response.Data[0].URL}, nil
}

func (client *Client) SubmitVideo(ctx context.Context, request repository.VideoGenerationRequest) (repository.VideoGenerationResult, error) {
	payload := map[string]any{
		"request_id": request.RequestID, "model": request.Model, "prompt": request.Prompt,
		"size": request.Size, "fps": request.FPS, "quality": request.Quality,
		"with_audio": request.WithAudio, "watermark_enabled": true,
	}
	if request.ImageURL != "" {
		payload["image_url"] = request.ImageURL
	}
	var response struct {
		ID         string `json:"id"`
		TaskStatus string `json:"task_status"`
	}
	if err := client.doJSON(ctx, http.MethodPost, "/videos/generations", payload, &response); err != nil {
		return repository.VideoGenerationResult{}, err
	}
	if response.ID == "" {
		return repository.VideoGenerationResult{}, fmt.Errorf("zhipu video response contains no task ID")
	}
	return repository.VideoGenerationResult{
		ProviderTaskID: response.ID, Status: mapTaskStatus(response.TaskStatus),
	}, nil
}

func (client *Client) GetVideoResult(ctx context.Context, providerTaskID string) (repository.VideoGenerationResult, error) {
	var response struct {
		TaskStatus string `json:"task_status"`
		Video      []struct {
			URL      string `json:"url"`
			CoverURL string `json:"cover_image_url"`
		} `json:"video_result"`
	}
	if err := client.doJSON(ctx, http.MethodGet, "/async-result/"+providerTaskID, nil, &response); err != nil {
		return repository.VideoGenerationResult{}, err
	}
	result := repository.VideoGenerationResult{ProviderTaskID: providerTaskID, Status: mapTaskStatus(response.TaskStatus)}
	if len(response.Video) > 0 {
		result.URL = response.Video[0].URL
		result.CoverURL = response.Video[0].CoverURL
	}
	return result, nil
}

func (client *Client) doJSON(ctx context.Context, method string, path string, input any, output any) error {
	var body io.Reader
	if input != nil {
		encoded, err := json.Marshal(input)
		if err != nil {
			return err
		}
		body = bytes.NewReader(encoded)
	}
	request, err := http.NewRequestWithContext(ctx, method, client.baseURL+path, body)
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", "Bearer "+client.apiKey)
	request.Header.Set("Content-Type", "application/json")
	response, err := client.http.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		message, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		return fmt.Errorf("zhipu returned %d: %s", response.StatusCode, strings.TrimSpace(string(message)))
	}
	if err := json.NewDecoder(response.Body).Decode(output); err != nil {
		return fmt.Errorf("decode zhipu response: %w", err)
	}
	return nil
}

func mapTaskStatus(status string) string {
	switch status {
	case "SUCCESS":
		return repository.GenerationSucceeded
	case "FAIL":
		return repository.GenerationFailed
	default:
		return repository.GenerationRunning
	}
}
