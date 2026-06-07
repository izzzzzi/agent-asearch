package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	APIKeys  map[string]string `json:"api_keys,omitempty"`
	Defaults Defaults          `json:"defaults,omitempty"`
}

type Defaults struct {
	Sources []string `json:"sources,omitempty"`
	Limit   int      `json:"limit,omitempty"`
}

func Path() string {
	d := os.Getenv("ASEARCH_STATE_DIR")
	if d == "" {
		home, _ := os.UserHomeDir()
		d = filepath.Join(home, ".asearch")
	}
	return filepath.Join(d, "config.json")
}

func Load() (*Config, error) {
	p := Path()
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{
				APIKeys:  make(map[string]string),
				Defaults: Defaults{Limit: 20, Sources: []string{"web", "hn", "reddit"}},
			}, nil
		}
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.APIKeys == nil {
		cfg.APIKeys = make(map[string]string)
	}
	if cfg.Defaults.Limit == 0 {
		cfg.Defaults.Limit = 20
	}
	if cfg.Defaults.Sources == nil {
		cfg.Defaults.Sources = []string{"web", "hn", "reddit"}
	}
	return &cfg, nil
}

func Save(cfg *Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	p := Path()
	os.MkdirAll(filepath.Dir(p), 0700)
	return os.WriteFile(p, data, 0600)
}

func GetKey(name string) string {
	envKey := nameToEnv(name)
	if v := os.Getenv(envKey); v != "" {
		return v
	}
	cfg, err := Load()
	if err != nil {
		return ""
	}
	return cfg.APIKeys[name]
}

func SetKey(name, value string) error {
	cfg, err := Load()
	if err != nil {
		return err
	}
	cfg.APIKeys[name] = value
	return Save(cfg)
}

func GetDefaultLimit() int {
	cfg, err := Load()
	if err != nil {
		return 20
	}
	if cfg.Defaults.Limit <= 0 {
		return 20
	}
	return cfg.Defaults.Limit
}

func GetDefaultSources() []string {
	cfg, err := Load()
	if err != nil {
		return []string{"web", "hn", "reddit"}
	}
	if len(cfg.Defaults.Sources) == 0 {
		return []string{"web", "hn", "reddit"}
	}
	return cfg.Defaults.Sources
}

func nameToEnv(name string) string {
	switch name {
	case "tavily":
		return "TAVILY_API_KEY"
	case "exa":
		return "EXA_API_KEY"
	case "brave":
		return "BRAVE_API_KEY"
	case "serper":
		return "SERPER_API_KEY"
	case "serpapi":
		return "SERPAPI_API_KEY"
	case "perplexity":
		return "PERPLEXITY_API_KEY"
	case "you":
		return "YOU_API_KEY"
	case "firecrawl":
		return "FIRECRAWL_API_KEY"
	case "parallel":
		return "PARALLEL_API_KEY"
	case "jina":
		return "JINA_API_KEY"
	case "github":
		return "GITHUB_TOKEN"
	case "searxng":
		return "ASEARCH_SEARXNG_URL"
	case "twitter":
		return "TWITTER_BEARER_TOKEN"
	default:
		return ""
	}
}
