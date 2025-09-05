package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// HistoryConfig represents the saved download history
type HistoryConfig struct {
	URLs  []string `json:"urls"`
	Names []string `json:"names"`
}

// ConfigData represents the complete application configuration
type ConfigData struct {
	History HistoryConfig `json:"history"`
	Presets []Preset      `json:"presets"`
}

// getConfigDir returns the config directory path
func getConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	configDir := filepath.Join(homeDir, ".config", "babago")

	// Create directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", err
	}

	return configDir, nil
}

// getConfigFilePath returns the full path to the config file
func getConfigFilePath() (string, error) {
	configDir, err := getConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, "config.json"), nil
}

// SaveConfig saves the complete application configuration
func SaveConfig(urls []string, names []string, presets []Preset) error {
	filePath, err := getConfigFilePath()
	if err != nil {
		return err
	}

	// Ensure both slices have the same length
	if len(urls) != len(names) {
		// Pad with empty strings if needed
		for len(names) < len(urls) {
			names = append(names, "")
		}
		for len(urls) < len(names) {
			urls = append(urls, "")
		}
	}

	config := ConfigData{
		History: HistoryConfig{
			URLs:  urls,
			Names: names,
		},
		Presets: presets,
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0644)
}

// LoadConfig loads the complete application configuration
func LoadConfig() ([]string, []string, []Preset, error) {
	filePath, err := getConfigFilePath()
	if err != nil {
		return []string{}, []string{}, []Preset{}, err
	}

	// If file doesn't exist, return empty config
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return []string{}, []string{}, []Preset{}, nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return []string{}, []string{}, []Preset{}, err
	}

	var config ConfigData
	if err := json.Unmarshal(data, &config); err != nil {
		return []string{}, []string{}, []Preset{}, err
	}

	// Ensure both slices have the same length
	if len(config.History.URLs) != len(config.History.Names) {
		// Pad with empty strings if needed
		for len(config.History.Names) < len(config.History.URLs) {
			config.History.Names = append(config.History.Names, "")
		}
		for len(config.History.URLs) < len(config.History.Names) {
			config.History.URLs = append(config.History.URLs, "")
		}
	}

	return config.History.URLs, config.History.Names, config.Presets, nil
}

// AutoSaveConfig is a convenience function for saving complete config
func AutoSaveConfig(uv *URLView, pv *PresetsView) {
	if err := SaveConfig(uv.URLHistory, uv.HistoryNames, pv.Presets); err != nil {
		logToFile("Failed to save config: " + err.Error())
	} else {
		logToFile("Config saved successfully")
	}
}

// GetDefaultPresets returns the default preset configuration
func GetDefaultPresets() []Preset {
	return []Preset{
		{
			Name:   "Defaults",
			Active: true,
			Options: []Option{
				{Flag: "--format=best", Comment: "Download best quality", Enabled: true},
				{Flag: "--write-thumbnail", Comment: "Write thumbnail image", Enabled: true},
				{Flag: "--write-description", Comment: "Write video description", Enabled: true},
			},
		},
		{
			Name:   "For Music",
			Active: false,
			Options: []Option{
				{Flag: "--extract-audio", Comment: "Extract audio only", Enabled: true},
				{Flag: "--audio-format mp3", Comment: "Convert to MP3", Enabled: true},
				{Flag: "--audio-quality 0", Comment: "Best audio quality", Enabled: true},
				{Flag: "--embed-thumbnail", Comment: "Embed thumbnail in audio", Enabled: true},
			},
		},
		{
			Name:   "Low Quality",
			Active: false,
			Options: []Option{
				{Flag: "--format=best[height<=720]", Comment: "Max 720p video", Enabled: true},
				{Flag: "--write-sub", Comment: "Download subtitles", Enabled: true},
				{Flag: "--sub-lang en", Comment: "English subtitles", Enabled: true},
				{Flag: "--embed-subs", Comment: "Embed subtitles", Enabled: true},
			},
		},
	}
}
