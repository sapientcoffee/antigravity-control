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
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// UpdatePlugin updates an outdated plugin via git pull or re-clone.
func UpdatePlugin(ctx context.Context, item AuditResult) error {
	target := item.Path
	if resolved, err := filepath.EvalSymlinks(item.Path); err == nil {
		target = resolved
	}

	gitDir := filepath.Join(target, ".git")
	if fi, err := os.Stat(gitDir); err == nil && (fi.IsDir() || fi.Mode().IsRegular()) {
		// Git pull
		cmd := exec.CommandContext(ctx, "git", "-C", target, "pull", "--ff-only")
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("git pull failed for %s: %s (%w)", item.Name, string(out), err)
		}
		return nil
	}

	// Non-git repo with repository URL
	if item.RepoURL != "" {
		backupDir := target + ".bak"
		_ = os.RemoveAll(backupDir)

		if err := os.Rename(target, backupDir); err != nil {
			return fmt.Errorf("backing up existing plugin %s: %w", item.Name, err)
		}

		cmd := exec.CommandContext(ctx, "git", "clone", "--depth", "1", item.RepoURL, target)
		out, err := cmd.CombinedOutput()
		if err != nil {
			// Restore backup
			_ = os.Rename(backupDir, target)
			return fmt.Errorf("git clone failed for %s: %s (%w)", item.Name, string(out), err)
		}

		_ = os.RemoveAll(backupDir)
		return nil
	}

	return fmt.Errorf("plugin %s cannot be updated automatically (no git repo or repository URL)", item.Name)
}
