package service

import (
	"fmt"
	"sort"

	"agent-canvas/backend/internal/model"
)

type ModelRegistry struct {
	profiles map[string]model.ModelProfile
}

func (registry *ModelRegistry) List() []model.ModelProfile {
	profiles := make([]model.ModelProfile, 0, len(registry.profiles))
	for _, profile := range registry.profiles {
		profiles = append(profiles, profile)
	}
	sort.Slice(profiles, func(i, j int) bool { return profiles[i].ID < profiles[j].ID })
	return profiles
}

func NewModelRegistry(profiles []model.ModelProfile) *ModelRegistry {
	registry := &ModelRegistry{profiles: make(map[string]model.ModelProfile, len(profiles))}
	for _, profile := range profiles {
		registry.profiles[profile.ID] = profile
	}
	return registry
}

func (registry *ModelRegistry) Resolve(modelID string) (model.ModelProfile, error) {
	profile, ok := registry.profiles[modelID]
	if !ok {
		return model.ModelProfile{}, fmt.Errorf("unknown model %q", modelID)
	}
	return profile, nil
}

func DefaultZhipuModelRegistry() *ModelRegistry {
	return NewZhipuModelRegistry("glm-4.7-flash", "cogview-3-flash", "cogvideox-flash")
}

func NewZhipuModelRegistry(agentModel string, imageModel string, videoModel string) *ModelRegistry {
	return NewModelRegistry([]model.ModelProfile{
		{
			ID: "canvas-agent", Provider: "zhipu", UpstreamModel: agentModel, Capability: "agent",
			Defaults: map[string]any{"thinking": "enabled"},
		},
		{
			ID: "canvas-image-fast", Provider: "zhipu", UpstreamModel: imageModel, Capability: model.GenerationKindImage,
			Defaults: map[string]any{"size": "1024x1024", "quality": "standard"},
		},
		{
			ID: "canvas-video-fast", Provider: "zhipu", UpstreamModel: videoModel, Capability: model.GenerationKindVideo,
			Defaults: map[string]any{"size": "1920x1080", "fps": 30, "quality": "speed", "withAudio": false},
		},
	})
}
