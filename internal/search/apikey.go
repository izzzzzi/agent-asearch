package search

import "github.com/izzzzzi/agent-asearch/internal/config"

func apiKey(name string) string {
	return config.GetKey(name)
}
