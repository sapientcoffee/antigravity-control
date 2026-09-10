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
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sapientcoffee/antigravity-control/pkg/config"
	"github.com/sapientcoffee/antigravity-control/pkg/fsutil"
	"github.com/sapientcoffee/antigravity-control/pkg/persona"
)

func setupTestCLIEnvironment(t *testing.T) string {
	tmpDir := t.TempDir()
	paths, err := config.ResolvePaths(tmpDir)
	if err != nil {
		t.Fatalf("ResolvePaths failed: %v", err)
	}

	_ = fsutil.EnsureDir(paths.PersonasDir)
	_ = fsutil.EnsureDir(paths.CatalogDir)
	_ = fsutil.EnsureDir(paths.SkillsDir)

	_ = fsutil.EnsureDir(filepath.Join(paths.CatalogDir, "git-delivery"))
	_ = fsutil.EnsureDir(filepath.Join(paths.CatalogDir, "graphify"))

	corePersona := persona.Persona{
		Name:            "core",
		DisplayName:     "Minimalist Core",
		Description:     "Baseline core persona",
		EnabledPlugins:  []string{"antigravity-control"},
		DisabledPlugins: []string{"bean-to-cup"},
		ActiveSkills:    []string{"git-delivery"},
	}
	coreData, _ := json.Marshal(corePersona)
	_ = os.WriteFile(filepath.Join(paths.PersonasDir, "core.json"), coreData, 0644)

	return tmpDir
}

func executeCommand(args ...string) (string, error) {
	rootCmd := NewRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(args)

	err := rootCmd.Execute()
	return buf.String(), err
}

func TestVersionCmd(t *testing.T) {
	out, err := executeCommand("version")
	if err != nil {
		t.Fatalf("version failed: %v", err)
	}
	if !strings.Contains(out, "agyctl (Antigravity Control) v1.0.0") {
		t.Errorf("unexpected output: %s", out)
	}
}

func TestSwitchAndCurrentCmd(t *testing.T) {
	tmpDir := setupTestCLIEnvironment(t)

	// Test switch
	outSwitch, err := executeCommand("--gemini-dir", tmpDir, "switch", "core")
	if err != nil {
		t.Fatalf("switch failed: %v, out: %s", err, outSwitch)
	}
	if !strings.Contains(outSwitch, "switched to persona") || !strings.Contains(outSwitch, "Minimalist Core") {
		t.Errorf("unexpected switch output: %s", outSwitch)
	}

	// Test current
	outCurrent, err := executeCommand("--gemini-dir", tmpDir, "current")
	if err != nil {
		t.Fatalf("current failed: %v, out: %s", err, outCurrent)
	}
	if !strings.Contains(outCurrent, "Active Persona") || !strings.Contains(outCurrent, "Minimalist Core") {
		t.Errorf("unexpected current output: %s", outCurrent)
	}

	// Test list
	outList, err := executeCommand("--gemini-dir", tmpDir, "list")
	if err != nil {
		t.Fatalf("list failed: %v, out: %s", err, outList)
	}
	if !strings.Contains(outList, "* (active)") || !strings.Contains(outList, "core") {
		t.Errorf("unexpected list output: %s", outList)
	}

	// Test reset
	outReset, err := executeCommand("--gemini-dir", tmpDir, "reset")
	if err != nil {
		t.Fatalf("reset failed: %v, out: %s", err, outReset)
	}
	if !strings.Contains(outReset, "switched to persona") {
		t.Errorf("unexpected reset output: %s", outReset)
	}

	// Test persona subcommands (e.g. persona current)
	outPersonaCur, err := executeCommand("--gemini-dir", tmpDir, "persona", "current")
	if err != nil {
		t.Fatalf("persona current failed: %v", err)
	}
	if !strings.Contains(outPersonaCur, "Active Persona") || !strings.Contains(outPersonaCur, "Minimalist Core") {
		t.Errorf("unexpected persona current output: %s", outPersonaCur)
	}
}

func TestPluginAndHookCmd(t *testing.T) {
	tmpDir := setupTestCLIEnvironment(t)
	paths, err := config.ResolvePaths(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	// Create a mock plugin
	mockDir := filepath.Join(paths.PluginsDir, "test-plugin")
	_ = fsutil.EnsureDir(mockDir)
	mockManifest := `{"name": "test-plugin", "version": "1.0.0", "description": "Mock plugin"}`
	_ = os.WriteFile(filepath.Join(mockDir, "plugin.json"), []byte(mockManifest), 0644)

	// Test plugin list
	outList, err := executeCommand("--gemini-dir", tmpDir, "plugin", "list")
	if err != nil {
		t.Fatalf("plugin list failed: %v, out: %s", err, outList)
	}
	if !strings.Contains(outList, "test-plugin") || !strings.Contains(outList, "1.0.0") {
		t.Errorf("unexpected plugin list output: %s", outList)
	}

	// Test check --json
	outCheckJSON, err := executeCommand("--gemini-dir", tmpDir, "check", "--json")
	if err != nil {
		t.Fatalf("check --json failed: %v, out: %s", err, outCheckJSON)
	}
	if !strings.Contains(outCheckJSON, `"name": "test-plugin"`) {
		t.Errorf("unexpected check json output: %s", outCheckJSON)
	}

	// Test hook session-start
	outHook, err := executeCommand("--gemini-dir", tmpDir, "hook", "session-start")
	if err != nil {
		t.Fatalf("hook session-start failed: %v, out: %s", err, outHook)
	}
	// With mock plugin up to date (no remote), outHook should be empty
	if strings.Contains(outHook, "Error") {
		t.Errorf("unexpected hook error: %s", outHook)
	}

	// Verify timestamp file was generated
	if _, err := os.Stat(paths.HookTimestampFile); os.IsNotExist(err) {
		t.Errorf("expected hook timestamp file %s to exist", paths.HookTimestampFile)
	}
}
