package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	openai "github.com/sashabaranov/go-openai"
	"gopkg.in/yaml.v3"
)

type PromptConfig struct {
	openai.ChatCompletionRequest `yaml:",inline"`
	ResponseFormat               `yaml:",inline"`
}

type ResponseFormat struct {
	Type       openai.ChatCompletionResponseFormatType `json:"type,omitempty"`
	JSONSchema *ResponseFormatJSONSchema               `json:"json_schema,omitempty"`
}

type ResponseFormatJSONSchema struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Schema      any    `json:"schema"`
	Strict      bool   `json:"strict"`
}

func (pc *PromptConfig) UnmarshalJSON(data []byte) error {
	type Alias PromptConfig
	aux := &struct {
		*Alias
		ResponseFormat json.RawMessage `json:"response_format,omitempty"`
	}{
		Alias: (*Alias)(pc),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if len(aux.ResponseFormat) > 0 {
		return json.Unmarshal(aux.ResponseFormat, &pc.ResponseFormat)
	}

	return nil
}

func loadPromptConfig(filename string) (PromptConfig, error) {
	var config PromptConfig

	data, err := os.ReadFile(filename)
	if err != nil {
		return config, fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".yaml", ".yml":
		yamlConfig := map[string]any{}
		err = yaml.Unmarshal(data, &yamlConfig)
		if err != nil {
			return config, err
		}

		bs, err := json.Marshal(yamlConfig)
		if err != nil {
			return config, err
		}

		err = json.Unmarshal(bs, &config)
		if err != nil {
			return config, err
		}

	case ".json":
		err = json.Unmarshal(data, &config)
		if err != nil {
			return config, err
		}
	default:
		return config, fmt.Errorf("unsupported file format: %s (use .json, .yaml, or .yml)", ext)
	}

	return config, nil
}
