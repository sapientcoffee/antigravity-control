// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"path/filepath"
	"testing"
)

func TestResolvePaths(t *testing.T) {
	tmpDir := t.TempDir()
	paths, err := ResolvePaths(tmpDir)
	if err != nil {
		t.Fatalf("ResolvePaths failed: %v", err)
	}

	if paths.GeminiHome != tmpDir {
		t.Errorf("got GeminiHome=%s, want %s", paths.GeminiHome, tmpDir)
	}
	expectedConfig := filepath.Join(tmpDir, "config", "config.json")
	if paths.ConfigFile != expectedConfig {
		t.Errorf("got ConfigFile=%s, want %s", paths.ConfigFile, expectedConfig)
	}
}

func TestLoadSaveConfigJSON(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.json")

	// Missing file should return default structure
	cfg, err := LoadConfigJSON(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfigJSON failed on missing file: %v", err)
	}
	plugins, ok := cfg["plugins"].(map[string]any)
	if !ok {
		t.Fatalf("expected plugins map, got %T", cfg["plugins"])
	}

	// Add plugin and user settings
	plugins["test-plugin"] = map[string]any{"enabled": true}
	cfg["userSettings"] = map[string]any{"theme": "dark"}

	if err := SaveConfigJSON(cfgPath, cfg); err != nil {
		t.Fatalf("SaveConfigJSON failed: %v", err)
	}

	reloaded, err := LoadConfigJSON(cfgPath)
	if err != nil {
		t.Fatalf("reloading config failed: %v", err)
	}
	reloadedPlugins := reloaded["plugins"].(map[string]any)
	testPlug := reloadedPlugins["test-plugin"].(map[string]any)
	if testPlug["enabled"] != true {
		t.Errorf("expected test-plugin enabled=true, got %v", testPlug["enabled"])
	}
	userSettings := reloaded["userSettings"].(map[string]any)
	if userSettings["theme"] != "dark" {
		t.Errorf("expected theme=dark, got %v", userSettings["theme"])
	}
}

func TestActivePersona(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, ".active_persona")

	// Default should be core
	if got := GetActivePersona(stateFile); got != "core" {
		t.Errorf("got %s, want core", got)
	}

	if err := SetActivePersona(stateFile, "architect"); err != nil {
		t.Fatalf("SetActivePersona failed: %v", err)
	}

	if got := GetActivePersona(stateFile); got != "architect" {
		t.Errorf("got %s, want architect", got)
	}
}
