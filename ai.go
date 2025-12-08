package main

import (
	"encoding/json"
)

// AIConfig AI 配置
type AIConfig struct {
	BaseURL string `json:"baseURL"`
	APIKey  string `json:"apiKey"`
	Model   string `json:"model"`
}

// GetAIConfig 获取 AI 配置
func (a *App) GetAIConfig() (*AIConfig, error) {
	store, err := a.countdownService.loadStore()
	if err != nil {
		return nil, err
	}

	if configData, ok := store["ai_config"]; ok {
		bytes, _ := json.Marshal(configData)
		var config AIConfig
		if err := json.Unmarshal(bytes, &config); err == nil {
			return &config, nil
		}
	}

	// 返回空配置
	return &AIConfig{}, nil
}

// SaveAIConfig 保存 AI 配置
func (a *App) SaveAIConfig(config AIConfig) error {
	store, err := a.countdownService.loadStore()
	if err != nil {
		return err
	}

	store["ai_config"] = config
	return a.countdownService.saveStore(store)
}
