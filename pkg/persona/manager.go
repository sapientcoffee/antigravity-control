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

package persona

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/sapientcoffee/antigravity-control/pkg/config"
	"github.com/sapientcoffee/antigravity-control/pkg/fsutil"
)

// GetAvailablePersonas scans paths.PersonasDir (and falls back to paths.RepoPersonasDir)
// for persona JSON definition files.
func GetAvailablePersonas(paths *config.Paths) (map[string]Persona, error) {
	personas := make(map[string]Persona)

	// Helper to load directory of JSON profiles
	loadDir := func(dir string) {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
				continue
			}
			filePath := filepath.Join(dir, entry.Name())
			data, err := os.ReadFile(filePath)
			if err != nil {
				continue
			}
			var p Persona
			if err := json.Unmarshal(data, &p); err != nil {
				continue
			}
			if p.Name == "" {
				p.Name = strings.TrimSuffix(entry.Name(), ".json")
			}
			if p.DisplayName == "" {
				p.DisplayName = p.Name
			}
			if absPath, err := filepath.Abs(filePath); err == nil {
				p.ConfigPath = absPath
			} else {
				p.ConfigPath = filePath
			}
			personas[p.Name] = p
		}
	}

	// 1. Try paths.PersonasDir (~/.gemini/personas)
	if paths.PersonasDir != "" {
		loadDir(paths.PersonasDir)
	}

	// 2. If empty, check repo personas directory as fallback
	if len(personas) == 0 && paths.RepoPersonasDir != "" {
		loadDir(paths.RepoPersonasDir)
	}

	return personas, nil
}

// Switch changes the active persona by toggling plugins in config.json,
// atomically refreshing symlinks in ~/.gemini/config/skills/, and recording state.
func Switch(paths *config.Paths, personaName string) (*SwitchResult, error) {
	personas, err := GetAvailablePersonas(paths)
	if err != nil {
		return nil, fmt.Errorf("loading personas: %w", err)
	}

	targetPersona, ok := personas[personaName]
	if !ok {
		var available []string
		for k := range personas {
			available = append(available, k)
		}
		sort.Strings(available)
		return nil, fmt.Errorf("unknown persona '%s'. Available: %s", personaName, strings.Join(available, ", "))
	}

	// 1. Update plugins in config.json
	cfg, err := config.LoadConfigJSON(paths.ConfigFile)
	if err != nil {
		return nil, fmt.Errorf("loading config.json: %w", err)
	}

	plugins, ok := cfg["plugins"].(map[string]any)
	if !ok {
		plugins = make(map[string]any)
		cfg["plugins"] = plugins
	}

	for _, plugin := range targetPersona.EnabledPlugins {
		if cur, exists := plugins[plugin].(map[string]any); exists {
			cur["enabled"] = true
		} else {
			plugins[plugin] = map[string]any{"enabled": true}
		}
	}

	for _, plugin := range targetPersona.DisabledPlugins {
		if cur, exists := plugins[plugin].(map[string]any); exists {
			cur["enabled"] = false
		} else {
			plugins[plugin] = map[string]any{"enabled": false}
		}
	}

	if err := config.SaveConfigJSON(paths.ConfigFile, cfg); err != nil {
		return nil, fmt.Errorf("saving config.json: %w", err)
	}

	// 2. Ensure catalog and skills directories exist
	if err := fsutil.EnsureDir(paths.CatalogDir); err != nil {
		return nil, fmt.Errorf("ensuring catalog directory: %w", err)
	}
	if err := fsutil.EnsureDir(paths.SkillsDir); err != nil {
		return nil, fmt.Errorf("ensuring skills directory: %w", err)
	}

	// 3. Clean existing symlinks in paths.SkillsDir (preserve physical non-symlink directories)
	existingEntries, err := os.ReadDir(paths.SkillsDir)
	if err == nil {
		for _, entry := range existingEntries {
			entryPath := filepath.Join(paths.SkillsDir, entry.Name())
			_, _ = fsutil.RemoveSymlink(entryPath)
		}
	}

	// 4. Create symlinks for active_skills from CatalogDir or FallbackSkillsDir
	linkedCount := 0
	var missingSkills []string

	for _, skill := range targetPersona.ActiveSkills {
		catalogTarget := filepath.Join(paths.CatalogDir, skill)
		fallbackTarget := filepath.Join(paths.FallbackSkillsDir, skill)

		var target string
		if fi, err := os.Stat(catalogTarget); err == nil && fi.IsDir() {
			target = catalogTarget
		} else if fi, err := os.Stat(fallbackTarget); err == nil && fi.IsDir() {
			target = fallbackTarget
		}

		if target != "" {
			dest := filepath.Join(paths.SkillsDir, skill)
			if err := fsutil.CreateOrUpdateSymlink(target, dest); err == nil {
				linkedCount++
			}
		} else {
			missingSkills = append(missingSkills, skill)
		}
	}

	// 5. Record state
	if err := config.SetActivePersona(paths.StateFile, personaName); err != nil {
		// Log or return error if state file cannot be written
		return nil, fmt.Errorf("updating active persona state: %w", err)
	}

	return &SwitchResult{
		Persona:        targetPersona,
		EnabledPlugins: targetPersona.EnabledPlugins,
		LinkedCount:    linkedCount,
		MissingSkills:  missingSkills,
	}, nil
}

// GetCurrent retrieves the current persona status including loaded skills and active plugins.
func GetCurrent(paths *config.Paths) (*CurrentStatus, error) {
	activeName := config.GetActivePersona(paths.StateFile)
	personas, _ := GetAvailablePersonas(paths)

	var activePersona *Persona
	if p, ok := personas[activeName]; ok {
		activePersona = &p
	}

	var activeSkills []string
	if entries, err := os.ReadDir(paths.SkillsDir); err == nil {
		for _, e := range entries {
			activeSkills = append(activeSkills, e.Name())
		}
		sort.Strings(activeSkills)
	}

	var activePlugins []string
	if cfg, err := config.LoadConfigJSON(paths.ConfigFile); err == nil {
		if plugins, ok := cfg["plugins"].(map[string]any); ok {
			for name, val := range plugins {
				if m, ok := val.(map[string]any); ok {
					if enabled, ok := m["enabled"].(bool); ok && enabled {
						activePlugins = append(activePlugins, name)
					}
				}
			}
			sort.Strings(activePlugins)
		}
	}

	return &CurrentStatus{
		ActivePersonaName: activeName,
		Persona:           activePersona,
		ActiveSkills:      activeSkills,
		ActivePlugins:     activePlugins,
	}, nil
}

// LoadItem temporarily enables a plugin or links a skill into the active session.
func LoadItem(paths *config.Paths, itemName string) (ItemType, error) {
	// 1. Check if it's a plugin in paths.PluginsDir
	pluginPath := filepath.Join(paths.PluginsDir, itemName)
	if fi, err := os.Stat(pluginPath); err == nil && fi.IsDir() {
		cfg, err := config.LoadConfigJSON(paths.ConfigFile)
		if err != nil {
			return "", fmt.Errorf("loading config.json: %w", err)
		}
		plugins, ok := cfg["plugins"].(map[string]any)
		if !ok {
			plugins = make(map[string]any)
			cfg["plugins"] = plugins
		}
		if cur, ok := plugins[itemName].(map[string]any); ok {
			cur["enabled"] = true
		} else {
			plugins[itemName] = map[string]any{"enabled": true}
		}
		if err := config.SaveConfigJSON(paths.ConfigFile, cfg); err != nil {
			return "", fmt.Errorf("saving config.json: %w", err)
		}
		return ItemPlugin, nil
	}

	// 2. Check if it's a skill in CatalogDir or FallbackSkillsDir
	catalogTarget := filepath.Join(paths.CatalogDir, itemName)
	fallbackTarget := filepath.Join(paths.FallbackSkillsDir, itemName)

	var target string
	if fi, err := os.Stat(catalogTarget); err == nil && fi.IsDir() {
		target = catalogTarget
	} else if fi, err := os.Stat(fallbackTarget); err == nil && fi.IsDir() {
		target = fallbackTarget
	}

	if target != "" {
		dest := filepath.Join(paths.SkillsDir, itemName)
		if err := fsutil.CreateOrUpdateSymlink(target, dest); err != nil {
			return "", fmt.Errorf("symlinking skill %s: %w", itemName, err)
		}
		return ItemSkill, nil
	}

	return "", fmt.Errorf("could not find plugin or skill named '%s'", itemName)
}

// UnloadItem unloads an individual skill or disables a plugin.
func UnloadItem(paths *config.Paths, itemName string) (ItemType, error) {
	var itemType ItemType
	found := false

	// Check if in plugins
	cfg, err := config.LoadConfigJSON(paths.ConfigFile)
	if err == nil {
		if plugins, ok := cfg["plugins"].(map[string]any); ok {
			if cur, ok := plugins[itemName].(map[string]any); ok {
				cur["enabled"] = false
				_ = config.SaveConfigJSON(paths.ConfigFile, cfg)
				itemType = ItemPlugin
				found = true
			}
		}
	}

	// Check if in skills
	dest := filepath.Join(paths.SkillsDir, itemName)
	isLink, err := fsutil.IsSymlink(dest)
	if err == nil && isLink {
		if _, err := fsutil.RemoveSymlink(dest); err == nil {
			itemType = ItemSkill
			found = true
		}
	} else if _, err := os.Stat(dest); err == nil {
		return "", fmt.Errorf("'%s' is a physical directory; not removing to protect data", itemName)
	}

	if !found {
		return "", fmt.Errorf("could not find active plugin or skill named '%s'", itemName)
	}

	return itemType, nil
}
