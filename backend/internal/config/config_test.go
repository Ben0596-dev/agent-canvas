package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadYAMLConfiguration(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "config.yaml")
	contents := []byte(`
app:
  environment: test
http:
  addr: ":19090"
database:
  url: postgres://user:pass@postgres:5432/test?sslmode=disable
  allow_no_db: false
agent:
  provider: zhipu
  openai_compatible:
    base_url: https://api.openai.com/v1
    api_key: ""
    model: ""
zhipu:
  base_url: https://open.bigmodel.cn/api/paas/v4
  api_key: secret
  models:
    agent: glm-4.7-flash
    image: cogview-3-flash
    video: cogvideox-flash
`)
	if err := os.WriteFile(path, contents, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	config, err := Load(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if config.HTTPAddr != ":19090" || config.AgentProvider != "zhipu" {
		t.Fatalf("unexpected config: %+v", config)
	}
	if config.ZhipuAPIKey != "secret" || config.ZhipuVideoModel != "cogvideox-flash" {
		t.Fatalf("zhipu config was not loaded")
	}
}

func TestLoadRejectsUnknownYAMLFields(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "config.yaml")
	if err := os.WriteFile(path, []byte("unknown: true\n"), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("expected unknown field error")
	}
}
