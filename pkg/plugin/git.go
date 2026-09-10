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
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type GitInfo struct {
	Commit    string
	RemoteURL string
	Path      string
}

// GetGitInfo inspects a directory for git repository status, returning commit and remote URL if present.
func GetGitInfo(ctx context.Context, dir string) *GitInfo {
	target := dir
	if fi, err := os.Lstat(dir); err == nil && fi.Mode()&os.ModeSymlink != 0 {
		if resolved, err := filepath.EvalSymlinks(dir); err == nil {
			target = resolved
		}
	}

	gitDir := filepath.Join(target, ".git")
	if fi, err := os.Stat(gitDir); err != nil || (!fi.IsDir() && !fi.Mode().IsRegular()) {
		return nil
	}

	commitCmd := exec.CommandContext(ctx, "git", "-C", target, "rev-parse", "HEAD")
	commitOut, err := commitCmd.Output()
	if err != nil {
		return nil
	}
	commit := strings.TrimSpace(string(commitOut))

	remoteCmd := exec.CommandContext(ctx, "git", "-C", target, "config", "--get", "remote.origin.url")
	remoteOut, _ := remoteCmd.Output()
	remoteURL := strings.TrimSpace(string(remoteOut))

	return &GitInfo{
		Commit:    commit,
		RemoteURL: remoteURL,
		Path:      target,
	}
}

// GetRemoteGitHead fetches the latest remote commit SHA for HEAD via git ls-remote.
func GetRemoteGitHead(ctx context.Context, remoteURL string) string {
	if remoteURL == "" {
		return ""
	}

	subCtx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()

	cmd := exec.CommandContext(subCtx, "git", "ls-remote", remoteURL, "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}

	fields := strings.Fields(string(out))
	if len(fields) > 0 {
		return fields[0]
	}
	return ""
}
