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

package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sapientcoffee/antigravity-control/pkg/config"
	"github.com/sapientcoffee/antigravity-control/pkg/fsutil"
	"github.com/sapientcoffee/antigravity-control/pkg/persona"
)

func TestEndToEndLifecycle(t *testing.T) {
	sandbox := t.TempDir()
	paths, err := config.ResolvePaths(sandbox)
	if err != nil {
		t.Fatalf("ResolvePaths failed: %v", err)
	}

	_ = fsutil.EnsureDir(paths.PersonasDir)
	_ = fsutil.EnsureDir(paths.CatalogDir)
	_ = fsutil.EnsureDir(paths.SkillsDir)
	_ = fsutil.EnsureDir(paths.PluginsDir)

	// Populate Catalog
	_ = fsutil.EnsureDir(filepath.Join(paths.CatalogDir, "git-delivery"))
	_ = fsutil.EnsureDir(filepath.Join(paths.CatalogDir, "graphify"))
	_ = fsutil.EnsureDir(filepath.Join(paths.CatalogDir, "ideator"))

	// Write personas
	coreP := persona.Persona{
		Name:            "core",
		DisplayName:     "Minimalist Core",
		Description:     "Core baseline",
		EnabledPlugins:  []string{"antigravity-control"},
		DisabledPlugins: []string{"test-plugin"},
		ActiveSkills:    []string{"git-delivery"},
	}
	archP := persona.Persona{
		Name:            "architect",
		DisplayName:     "Architect",
		Description:     "Architecture suite",
		EnabledPlugins:  []string{"antigravity-control", "test-plugin"},
		DisabledPlugins: []string{},
		ActiveSkills:    []string{"ideator"},
	}
	cData, _ := json.Marshal(coreP)
	aData, _ := json.Marshal(archP)
	_ = os.WriteFile(filepath.Join(paths.PersonasDir, "core.json"), cData, 0644)
	_ = os.WriteFile(filepath.Join(paths.PersonasDir, "architect.json"), aData, 0644)

	// Mock plugin
	plugDir := filepath.Join(paths.PluginsDir, "test-plugin")
	_ = fsutil.EnsureDir(plugDir)
	plugManifest := `{"name": "test-plugin", "version": "1.0.0", "description": "Test plugin"}`
	_ = os.WriteFile(filepath.Join(plugDir, "plugin.json"), []byte(plugManifest), 0644)

	// 1. Initial switch to core
	out, err := executeCommand("--gemini-dir", sandbox, "switch", "core")
	if err != nil {
		t.Fatalf("switch core failed: %v, out: %s", err, out)
	}
	if !strings.Contains(out, "Minimalist Core") {
		t.Errorf("expected Minimalist Core in switch output, got: %s", out)
	}
	isLink, _ := fsutil.IsSymlink(filepath.Join(paths.SkillsDir, "git-delivery"))
	if !isLink {
		t.Errorf("expected git-delivery to be symlinked in skills dir")
	}

	// 2. Current status
	outCur, err := executeCommand("--gemini-dir", sandbox, "current")
	if err != nil {
		t.Fatalf("current failed: %v, out: %s", err, outCur)
	}
	if !strings.Contains(outCur, "core") {
		t.Errorf("expected core in current output: %s", outCur)
	}

	// 3. Switch to architect
	outArch, err := executeCommand("--gemini-dir", sandbox, "switch", "architect")
	if err != nil {
		t.Fatalf("switch architect failed: %v, out: %s", err, outArch)
	}
	if !strings.Contains(outArch, "Architect") {
		t.Errorf("expected Architect in switch output: %s", outArch)
	}
	// Verify git-delivery removed and ideator linked
	if isGitLink, _ := fsutil.IsSymlink(filepath.Join(paths.SkillsDir, "git-delivery")); isGitLink {
		t.Errorf("git-delivery should have been unlinked")
	}
	if isIdeatorLink, _ := fsutil.IsSymlink(filepath.Join(paths.SkillsDir, "ideator")); !isIdeatorLink {
		t.Errorf("ideator should have been symlinked")
	}

	// 4. Load ad-hoc skill
	outLoad, err := executeCommand("--gemini-dir", sandbox, "load", "graphify")
	if err != nil {
		t.Fatalf("load failed: %v, out: %s", err, outLoad)
	}
	if !strings.Contains(outLoad, "Skill loaded") || !strings.Contains(outLoad, "graphify") {
		t.Errorf("unexpected load output: %s", outLoad)
	}
	if isGraphLink, _ := fsutil.IsSymlink(filepath.Join(paths.SkillsDir, "graphify")); !isGraphLink {
		t.Errorf("graphify should be linked after load")
	}

	// 5. Unload ad-hoc skill
	outUnload, err := executeCommand("--gemini-dir", sandbox, "unload", "graphify")
	if err != nil {
		t.Fatalf("unload failed: %v, out: %s", err, outUnload)
	}
	if !strings.Contains(outUnload, "Skill unloaded") || !strings.Contains(outUnload, "graphify") {
		t.Errorf("unexpected unload output: %s", outUnload)
	}

	// 6. Reset to core
	outReset, err := executeCommand("--gemini-dir", sandbox, "reset")
	if err != nil {
		t.Fatalf("reset failed: %v, out: %s", err, outReset)
	}
	if !strings.Contains(outReset, "core") {
		t.Errorf("expected core in reset output: %s", outReset)
	}

	// 7. Hook session-start
	outHook, err := executeCommand("--gemini-dir", sandbox, "hook", "session-start")
	if err != nil {
		t.Fatalf("hook session-start failed: %v, out: %s", err, outHook)
	}
	if _, err := os.Stat(paths.HookTimestampFile); os.IsNotExist(err) {
		t.Errorf("expected timestamp file to exist")
	}
}
