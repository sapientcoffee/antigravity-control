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
	"os"
	"path/filepath"
	"testing"

	"github.com/sapientcoffee/antigravity-control/pkg/config"
	"github.com/sapientcoffee/antigravity-control/pkg/fsutil"
)

func setupTestEnvironment(t *testing.T) (*config.Paths, string) {
	tmpDir := t.TempDir()
	paths, err := config.ResolvePaths(tmpDir)
	if err != nil {
		t.Fatalf("ResolvePaths failed: %v", err)
	}

	// Create directories
	_ = fsutil.EnsureDir(paths.PersonasDir)
	_ = fsutil.EnsureDir(paths.CatalogDir)
	_ = fsutil.EnsureDir(paths.SkillsDir)
	_ = fsutil.EnsureDir(paths.PluginsDir)

	// Create dummy skills in catalog
	_ = fsutil.EnsureDir(filepath.Join(paths.CatalogDir, "git-delivery"))
	_ = fsutil.EnsureDir(filepath.Join(paths.CatalogDir, "graphify"))
	_ = fsutil.EnsureDir(filepath.Join(paths.CatalogDir, "ideator"))

	// Create sample core persona
	corePersona := Persona{
		Name:            "core",
		DisplayName:     "Minimalist Core",
		Description:     "Core baseline",
		EnabledPlugins:  []string{"antigravity-control"},
		DisabledPlugins: []string{"bean-to-cup"},
		ActiveSkills:    []string{"git-delivery", "graphify"},
	}
	coreData, _ := json.Marshal(corePersona)
	_ = os.WriteFile(filepath.Join(paths.PersonasDir, "core.json"), coreData, 0644)

	// Create sample architect persona
	archPersona := Persona{
		Name:            "architect",
		DisplayName:     "System Architect",
		Description:     "SDLC Tools",
		EnabledPlugins:  []string{"antigravity-control", "bean-to-cup"},
		DisabledPlugins: []string{"stitch-build"},
		ActiveSkills:    []string{"ideator"},
	}
	archData, _ := json.Marshal(archPersona)
	_ = os.WriteFile(filepath.Join(paths.PersonasDir, "architect.json"), archData, 0644)

	return paths, tmpDir
}

func TestGetAvailablePersonas(t *testing.T) {
	paths, _ := setupTestEnvironment(t)
	personas, err := GetAvailablePersonas(paths)
	if err != nil {
		t.Fatalf("GetAvailablePersonas failed: %v", err)
	}

	if len(personas) != 2 {
		t.Fatalf("expected 2 personas, got %d", len(personas))
	}
	if _, ok := personas["core"]; !ok {
		t.Errorf("missing core persona")
	}
	if _, ok := personas["architect"]; !ok {
		t.Errorf("missing architect persona")
	}
}

func TestSwitchPersona(t *testing.T) {
	paths, _ := setupTestEnvironment(t)

	// Switch to core
	res, err := Switch(paths, "core")
	if err != nil {
		t.Fatalf("Switch to core failed: %v", err)
	}
	if res.LinkedCount != 2 {
		t.Errorf("expected 2 linked skills, got %d", res.LinkedCount)
	}

	// Verify symlinks exist in SkillsDir
	isLink1, _ := fsutil.IsSymlink(filepath.Join(paths.SkillsDir, "git-delivery"))
	isLink2, _ := fsutil.IsSymlink(filepath.Join(paths.SkillsDir, "graphify"))
	if !isLink1 || !isLink2 {
		t.Errorf("expected git-delivery and graphify to be symlinked")
	}

	// Verify state file
	if got := config.GetActivePersona(paths.StateFile); got != "core" {
		t.Errorf("got active persona %s, want core", got)
	}

	// Switch to architect
	resArch, err := Switch(paths, "architect")
	if err != nil {
		t.Fatalf("Switch to architect failed: %v", err)
	}
	if resArch.LinkedCount != 1 {
		t.Errorf("expected 1 linked skill, got %d", resArch.LinkedCount)
	}

	// git-delivery should no longer be linked
	_, errGit := os.Lstat(filepath.Join(paths.SkillsDir, "git-delivery"))
	if !os.IsNotExist(errGit) {
		t.Errorf("expected git-delivery to be removed from active skills")
	}

	// ideator should be linked
	isLinkArch, _ := fsutil.IsSymlink(filepath.Join(paths.SkillsDir, "ideator"))
	if !isLinkArch {
		t.Errorf("expected ideator to be symlinked")
	}

	// Verify plugins toggled in config.json
	cfg, err := config.LoadConfigJSON(paths.ConfigFile)
	if err != nil {
		t.Fatal(err)
	}
	plugins := cfg["plugins"].(map[string]any)
	bean := plugins["bean-to-cup"].(map[string]any)
	if bean["enabled"] != true {
		t.Errorf("expected bean-to-cup enabled=true, got %v", bean["enabled"])
	}
}

func TestGetCurrent(t *testing.T) {
	paths, _ := setupTestEnvironment(t)
	_, _ = Switch(paths, "core")

	status, err := GetCurrent(paths)
	if err != nil {
		t.Fatalf("GetCurrent failed: %v", err)
	}
	if status.ActivePersonaName != "core" {
		t.Errorf("got %s, want core", status.ActivePersonaName)
	}
	if len(status.ActiveSkills) != 2 {
		t.Errorf("expected 2 active skills, got %d", len(status.ActiveSkills))
	}
}

func TestLoadAndUnloadItem(t *testing.T) {
	paths, _ := setupTestEnvironment(t)
	_, _ = Switch(paths, "core")

	// 1. Load an unlinked skill
	itemType, err := LoadItem(paths, "ideator")
	if err != nil {
		t.Fatalf("LoadItem failed: %v", err)
	}
	if itemType != ItemSkill {
		t.Errorf("expected ItemSkill, got %v", itemType)
	}
	isLink, _ := fsutil.IsSymlink(filepath.Join(paths.SkillsDir, "ideator"))
	if !isLink {
		t.Errorf("expected ideator to be symlinked after load")
	}

	// 2. Unload the skill
	unloadedType, err := UnloadItem(paths, "ideator")
	if err != nil {
		t.Fatalf("UnloadItem failed: %v", err)
	}
	if unloadedType != ItemSkill {
		t.Errorf("expected ItemSkill, got %v", unloadedType)
	}
	_, errStat := os.Lstat(filepath.Join(paths.SkillsDir, "ideator"))
	if !os.IsNotExist(errStat) {
		t.Errorf("expected ideator symlink to be removed")
	}
}
