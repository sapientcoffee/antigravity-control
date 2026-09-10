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

package plugin

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/sapientcoffee/antigravity-control/pkg/config"
	"github.com/sapientcoffee/antigravity-control/pkg/fsutil"
)

func TestParseRepoOwnerName(t *testing.T) {
	cases := []struct {
		url       string
		wantOwner string
		wantRepo  string
		wantOK    bool
	}{
		{"https://github.com/sapientcoffee/antigravity-control", "sapientcoffee", "antigravity-control", true},
		{"https://github.com/sapientcoffee/antigravity-control.git", "sapientcoffee", "antigravity-control", true},
		{"git@github.com:owner/repo.git", "owner", "repo", true},
		{"https://gitlab.com/owner/repo", "", "", false},
		{"", "", "", false},
	}

	for _, c := range cases {
		owner, repo, ok := ParseRepoOwnerName(c.url)
		if ok != c.wantOK || owner != c.wantOwner || repo != c.wantRepo {
			t.Errorf("ParseRepoOwnerName(%q) = (%q, %q, %v); want (%q, %q, %v)",
				c.url, owner, repo, ok, c.wantOwner, c.wantRepo, c.wantOK)
		}
	}
}

func TestAuditPlugins(t *testing.T) {
	tmpDir := t.TempDir()
	paths, err := config.ResolvePaths(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	_ = fsutil.EnsureDir(paths.PluginsDir)
	testPluginDir := filepath.Join(paths.PluginsDir, "mock-plugin")
	_ = fsutil.EnsureDir(testPluginDir)

	manifest := Manifest{
		Name:        "mock-plugin",
		Version:     "1.0.0",
		Repository:  "https://github.com/foo/bar",
		Description: "A mock plugin for testing",
	}
	mData, _ := json.Marshal(manifest)
	_ = os.WriteFile(filepath.Join(testPluginDir, "plugin.json"), mData, 0644)

	ctx := context.Background()
	results, err := AuditPlugins(ctx, paths)
	if err != nil {
		t.Fatalf("AuditPlugins failed: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	res := results[0]
	if res.Name != "mock-plugin" {
		t.Errorf("got name %s, want mock-plugin", res.Name)
	}
	if res.LocalVersion != "1.0.0" {
		t.Errorf("got local_version %s, want 1.0.0", res.LocalVersion)
	}

	// Verify audit cache file was created
	if _, err := os.Stat(paths.AuditCacheFile); os.IsNotExist(err) {
		t.Errorf("expected audit cache file %s to exist", paths.AuditCacheFile)
	}
}
