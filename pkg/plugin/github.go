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
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"
)

var (
	githubURLRegex = regexp.MustCompile(`github\.com[/:]([A-Za-z0-9_.-]+)/([A-Za-z0-9_.-]+?)(?:\.git)?$`)
	cachedToken    string
	tokenOnce      sync.Once
)

type RemoteVersionInfo struct {
	Type   string // "release", "tag", "commit"
	Latest string
}

// ParseRepoOwnerName extracts the owner and repository name from a GitHub URL.
func ParseRepoOwnerName(url string) (owner, repo string, ok bool) {
	if url == "" {
		return "", "", false
	}
	url = strings.TrimRight(strings.TrimSpace(url), "/")
	matches := githubURLRegex.FindStringSubmatch(url)
	if len(matches) == 3 {
		return matches[1], matches[2], true
	}
	return "", "", false
}

// getGitHubToken retrieves an auth token from environment variables or gh CLI.
func getGitHubToken(ctx context.Context) string {
	tokenOnce.Do(func() {
		if token := os.Getenv("GITHUB_TOKEN"); token != "" {
			cachedToken = token
			return
		}
		if token := os.Getenv("GH_TOKEN"); token != "" {
			cachedToken = token
			return
		}

		// Try gh auth token with 1s timeout
		cmdCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
		defer cancel()
		cmd := exec.CommandContext(cmdCtx, "gh", "auth", "token")
		if out, err := cmd.Output(); err == nil {
			cachedToken = strings.TrimSpace(string(out))
		}
	})
	return cachedToken
}

// GetRemoteGitHubLatest checks GitHub for latest release tag, latest tag, or HEAD commit.
func GetRemoteGitHubLatest(ctx context.Context, owner, repo string) *RemoteVersionInfo {
	client := &http.Client{
		Timeout: 3 * time.Second,
	}
	token := getGitHubToken(ctx)

	makeReq := func(endpoint string) (*http.Response, error) {
		req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Accept", "application/vnd.github.v3+json")
		req.Header.Set("User-Agent", "agyctl/1.0.0")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		return client.Do(req)
	}

	// 1. Try releases/latest
	relURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)
	if resp, err := makeReq(relURL); err == nil {
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			var release struct {
				TagName string `json:"tag_name"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&release); err == nil && release.TagName != "" {
				return &RemoteVersionInfo{
					Type:   "release",
					Latest: strings.TrimPrefix(release.TagName, "v"),
				}
			}
		}
	}

	// 2. Try tags
	tagsURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/tags?per_page=1", owner, repo)
	if resp, err := makeReq(tagsURL); err == nil {
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			var tags []struct {
				Name string `json:"name"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&tags); err == nil && len(tags) > 0 && tags[0].Name != "" {
				return &RemoteVersionInfo{
					Type:   "tag",
					Latest: strings.TrimPrefix(tags[0].Name, "v"),
				}
			}
		}
	}

	// 3. Try commits/HEAD
	commitsURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/commits/HEAD", owner, repo)
	if resp, err := makeReq(commitsURL); err == nil {
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			var commit struct {
				SHA string `json:"sha"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&commit); err == nil && commit.SHA != "" {
				short := commit.SHA
				if len(short) > 7 {
					short = short[:7]
				}
				return &RemoteVersionInfo{
					Type:   "commit",
					Latest: short,
				}
			}
		}
	}

	// 4. Subprocess fallback using gh CLI
	ghCmd := exec.CommandContext(ctx, "gh", "api", fmt.Sprintf("repos/%s/%s/releases/latest", owner, repo), "--jq", ".tag_name")
	if out, err := ghCmd.Output(); err == nil && strings.TrimSpace(string(out)) != "" {
		return &RemoteVersionInfo{
			Type:   "release",
			Latest: strings.TrimPrefix(strings.TrimSpace(string(out)), "v"),
		}
	}

	return nil
}
