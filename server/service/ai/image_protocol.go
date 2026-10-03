package ai

import "strings"

// Fixed adapters keep image setup small without filtering model identifiers.
func imageRequestPayload(cfg ImageEndpointConfig, prompt, size string) (map[string]any, error) {
	payload := map[string]any{"model": cfg.Model, "prompt": prompt, "size": size}
	if cfg.Provider == "ark" {
		payload["size"] = map[string]string{"1024x1024": "2048x2048", "1536x1024": "2496x1664", "1024x1536": "1664x2496"}[size]
	} else {
		payload["n"] = 1
	}
	if !strings.HasPrefix(cfg.Model, "gpt-image") {
		payload["response_format"] = "b64_json"
	}
	return payload, nil
}
