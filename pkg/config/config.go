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
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sapientcoffee/antigravity-control/pkg/fsutil"
)

// Paths holds all resolved filesystem locations used by agyctl.
type Paths struct {
	GeminiHome           string
	ConfigDir            string
	ConfigFile           string
	SkillsDir            string
	CatalogDir           string
	FallbackSkillsDir    string
	PersonasDir          string
	RepoPersonasDir      string
	PluginsDir           string
	StateFile            string
	CacheDir             string
	AuditCacheFile       string
	HookTimestampFile    string
	HookCachedNoticeFile string
}

// ResolvePaths resolves all standard directories with optional override support.
func ResolvePaths(geminiHomeOverride string) (*Paths, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("resolving user home directory: %w", err)
	}

	geminiHome := geminiHomeOverride
	if geminiHome == "" {
		geminiHome = os.Getenv("GEMINI_HOME")
	}
	if geminiHome == "" {
		geminiHome = filepath.Join(home, ".gemini")
	}

	cacheBase := os.Getenv("XDG_CACHE_HOME")
	if cacheBase == "" {
		cacheBase = filepath.Join(home, ".cache")
	}
	cacheDir := filepath.Join(cacheBase, "antigravity-control")

	configDir := filepath.Join(geminiHome, "config")
	paths := &Paths{
		GeminiHome:           geminiHome,
		ConfigDir:            configDir,
		ConfigFile:           filepath.Join(configDir, "config.json"),
		SkillsDir:            filepath.Join(configDir, "skills"),
		CatalogDir:           filepath.Join(geminiHome, "catalog", "skills"),
		FallbackSkillsDir:    filepath.Join(home, ".agents", "skills"),
		PersonasDir:          filepath.Join(geminiHome, "personas"),
		RepoPersonasDir:      "personas",
		PluginsDir:           filepath.Join(configDir, "plugins"),
		StateFile:            filepath.Join(configDir, ".active_persona"),
		CacheDir:             cacheDir,
		AuditCacheFile:       filepath.Join(cacheDir, "update_audit.json"),
		HookTimestampFile:    filepath.Join(cacheDir, "last_check_ts"),
		HookCachedNoticeFile: filepath.Join(cacheDir, "cached_notice.md"),
	}

	return paths, nil
}

// LoadConfigJSON reads the Antigravity config.json file into a generic map,
// ensuring the "plugins" section is present as a map[string]any.
func LoadConfigJSON(path string) (map[string]any, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return map[string]any{
			"plugins": map[string]any{},
		}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config.json at %s: %w", path, err)
	}

	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("parsing config.json: %w", err)
	}
	if root == nil {
		root = make(map[string]any)
	}

	// Ensure "plugins" exists and is a map
	if _, ok := root["plugins"].(map[string]any); !ok {
		root["plugins"] = make(map[string]any)
	}

	return root, nil
}

// SaveConfigJSON atomically serializes the Antigravity config.json file with 2-space indentation.
func SaveConfigJSON(path string, data map[string]any) error {
	encoded, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding config.json: %w", err)
	}
	encoded = append(encoded, '\n')
	return fsutil.AtomicWriteFile(path, encoded, 0644)
}

// GetActivePersona reads the currently active persona name from stateFile, defaulting to "core".
func GetActivePersona(stateFile string) string {
	data, err := os.ReadFile(stateFile)
	if err == nil {
		val := strings.TrimSpace(string(data))
		if val != "" {
			return val
		}
	}
	return "core"
}

// SetActivePersona atomically records the currently active persona name.
func SetActivePersona(stateFile, name string) error {
	content := []byte(strings.TrimSpace(name) + "\n")
	return fsutil.AtomicWriteFile(stateFile, content, 0644)
}
