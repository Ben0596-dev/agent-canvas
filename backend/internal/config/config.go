package config

import (
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
)

type Config struct {
	HTTPAddr        string
	DatabaseURL     string
	Environment     string
	AllowNoDB       bool
	AgentProvider   string
	AgentBaseURL    string
	AgentAPIKey     string
	AgentModel      string
	ZhipuBaseURL    string
	ZhipuAPIKey     string
	ZhipuAgentModel string
	ZhipuImageModel string
	ZhipuVideoModel string
}

type fileConfig struct {
	App struct {
		Environment string `yaml:"environment"`
	} `yaml:"app"`
	HTTP struct {
		Addr string `yaml:"addr"`
	} `yaml:"http"`
	Database struct {
		URL       string `yaml:"url"`
		AllowNoDB bool   `yaml:"allow_no_db"`
	} `yaml:"database"`
	Agent struct {
		Provider         string                 `yaml:"provider"`
		OpenAICompatible openAICompatibleConfig `yaml:"openai_compatible"`
	} `yaml:"agent"`
	Zhipu zhipuConfig `yaml:"zhipu"`
}

type openAICompatibleConfig struct {
	BaseURL string `yaml:"base_url"`
	APIKey  string `yaml:"api_key"`
	Model   string `yaml:"model"`
}

type zhipuConfig struct {
	BaseURL string `yaml:"base_url"`
	APIKey  string `yaml:"api_key"`
	Models  struct {
		Agent string `yaml:"agent"`
		Image string `yaml:"image"`
		Video string `yaml:"video"`
	} `yaml:"models"`
}

func Load(path string) (Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return Config{}, fmt.Errorf("open config %q: %w", path, err)
	}
	defer file.Close()

	configuration := defaultFileConfig()
	if err := yaml.NewDecoder(file, yaml.DisallowUnknownField()).Decode(&configuration); err != nil {
		return Config{}, fmt.Errorf("decode config %q: %w", path, err)
	}
	config := Config{
		HTTPAddr:        configuration.HTTP.Addr,
		DatabaseURL:     configuration.Database.URL,
		Environment:     configuration.App.Environment,
		AllowNoDB:       configuration.Database.AllowNoDB,
		AgentProvider:   configuration.Agent.Provider,
		AgentBaseURL:    configuration.Agent.OpenAICompatible.BaseURL,
		AgentAPIKey:     configuration.Agent.OpenAICompatible.APIKey,
		AgentModel:      configuration.Agent.OpenAICompatible.Model,
		ZhipuBaseURL:    configuration.Zhipu.BaseURL,
		ZhipuAPIKey:     configuration.Zhipu.APIKey,
		ZhipuAgentModel: configuration.Zhipu.Models.Agent,
		ZhipuImageModel: configuration.Zhipu.Models.Image,
		ZhipuVideoModel: configuration.Zhipu.Models.Video,
	}
	if err := config.Validate(); err != nil {
		return Config{}, fmt.Errorf("validate config %q: %w", path, err)
	}
	return config, nil
}

func (config Config) Validate() error {
	if config.HTTPAddr == "" {
		return fmt.Errorf("http.addr is required")
	}
	if config.DatabaseURL == "" && !config.AllowNoDB {
		return fmt.Errorf("database.url is required when database.allow_no_db is false")
	}
	if config.AgentProvider != "local" && config.AgentProvider != "openai-compatible" && config.AgentProvider != "zhipu" {
		return fmt.Errorf("agent.provider must be local, zhipu, or openai-compatible")
	}
	if config.AgentProvider == "openai-compatible" && (config.AgentAPIKey == "" || config.AgentModel == "") {
		return fmt.Errorf("agent.openai_compatible.api_key and model are required")
	}
	if config.AgentProvider == "zhipu" && config.ZhipuAPIKey == "" {
		return fmt.Errorf("zhipu.api_key is required when agent.provider is zhipu")
	}
	return nil
}

func defaultFileConfig() fileConfig {
	var configuration fileConfig
	configuration.App.Environment = "development"
	configuration.HTTP.Addr = ":8080"
	configuration.Database.AllowNoDB = true
	configuration.Agent.Provider = "local"
	configuration.Agent.OpenAICompatible.BaseURL = "https://api.openai.com/v1"
	configuration.Zhipu.BaseURL = "https://open.bigmodel.cn/api/paas/v4"
	configuration.Zhipu.Models.Agent = "glm-4.7-flash"
	configuration.Zhipu.Models.Image = "cogview-3-flash"
	configuration.Zhipu.Models.Video = "cogvideox-flash"
	return configuration
}
