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
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestPersonaFilesValid(t *testing.T) {
	personasDir := filepath.Join("..", "personas")
	files, err := filepath.Glob(filepath.Join(personasDir, "*.json"))
	if err != nil {
		t.Fatalf("failed to glob personas: %v", err)
	}

	if len(files) < 5 {
		t.Errorf("expected at least 5 persona files, found %d", len(files))
	}

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Errorf("failed to read persona file %s: %v", file, err)
			continue
		}

		var m map[string]any
		if err := json.Unmarshal(data, &m); err != nil {
			t.Errorf("persona %s is invalid JSON: %v", file, err)
			continue
		}

		requiredKeys := []string{"name", "displayName", "enabled_plugins", "disabled_plugins", "active_skills"}
		for _, key := range requiredKeys {
			if _, exists := m[key]; !exists {
				t.Errorf("persona %s is missing required key %q", file, key)
			}
		}

		// Ensure antigravity-control is in enabled_plugins
		plugins, ok := m["enabled_plugins"].([]any)
		if !ok {
			t.Errorf("persona %s enabled_plugins is not an array", file)
			continue
		}
		foundControl := false
		for _, p := range plugins {
			if s, ok := p.(string); ok && s == "antigravity-control" {
				foundControl = true
				break
			}
		}
		if !foundControl {
			t.Errorf("persona %s missing 'antigravity-control' in enabled_plugins", file)
		}
	}
}

func TestAgyctlListCommand(t *testing.T) {
	out, err := executeCommand("list")
	if err != nil {
		t.Fatalf("executeCommand(list) failed: %v", err)
	}

	expected := []string{"core", "architect", "fullstack", "stitch"}
	for _, exp := range expected {
		if !strings.Contains(out, exp) {
			t.Errorf("expected %q in list output, got:\n%s", exp, out)
		}
	}
}

func TestAgyctlCheckJSONCommand(t *testing.T) {
	out, _ := executeCommand("check", "--json")
	var results []map[string]any
	if err := json.Unmarshal([]byte(out), &results); err != nil {
		t.Fatalf("expected valid JSON array from check --json, got: %s (err: %v)", out, err)
	}

	if len(results) == 0 {
		t.Errorf("expected at least 1 plugin result in check --json")
	}

	if _, ok := results[0]["name"]; !ok {
		t.Errorf("expected 'name' in check JSON item: %+v", results[0])
	}
	if _, ok := results[0]["status"]; !ok {
		t.Errorf("expected 'status' in check JSON item: %+v", results[0])
	}
}

func TestCheckUpdatesHookScript(t *testing.T) {
	hookScript := filepath.Join("..", "hooks", "check-updates.sh")
	if _, err := os.Stat(hookScript); err != nil {
		t.Skipf("hooks/check-updates.sh not found: %v", err)
	}

	cmd := exec.Command(hookScript)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("hook execution failed: %v, output: %s", err, string(out))
	}
}
