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
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/sapientcoffee/antigravity-control/pkg/config"
	"github.com/sapientcoffee/antigravity-control/pkg/fsutil"
)

// AuditPlugins audits all plugins in paths.PluginsDir concurrently using a worker pool.
func AuditPlugins(ctx context.Context, paths *config.Paths) ([]AuditResult, error) {
	if _, err := os.Stat(paths.PluginsDir); os.IsNotExist(err) {
		return []AuditResult{}, nil
	}

	entries, err := os.ReadDir(paths.PluginsDir)
	if err != nil {
		return nil, err
	}

	type target struct {
		name string
		path string
	}
	var targets []target

	for _, entry := range entries {
		p := filepath.Join(paths.PluginsDir, entry.Name())
		fi, err := os.Stat(p)
		if err != nil || !fi.IsDir() {
			continue
		}
		targets = append(targets, target{name: entry.Name(), path: p})
	}

	auditCtx, cancel := context.WithTimeout(ctx, 7*time.Second)
	defer cancel()

	results := make([]AuditResult, len(targets))
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 8) // max 8 concurrent worker goroutines

	for i, t := range targets {
		wg.Add(1)
		go func(idx int, tgt target) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			results[idx] = auditSinglePlugin(auditCtx, tgt.name, tgt.path)
		}(i, t)
	}

	wg.Wait()

	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})

	// Cache results
	if len(results) > 0 && paths.AuditCacheFile != "" {
		if cacheBytes, err := json.MarshalIndent(results, "", "  "); err == nil {
			_ = fsutil.AtomicWriteFile(paths.AuditCacheFile, append(cacheBytes, '\n'), 0644)
		}
	}

	return results, nil
}

func auditSinglePlugin(ctx context.Context, defaultName, dirPath string) AuditResult {
	manifestFile := filepath.Join(dirPath, "plugin.json")
	name := defaultName
	localVersion := "unknown"
	repoURL := ""

	if data, err := os.ReadFile(manifestFile); err == nil {
		var m Manifest
		if err := json.Unmarshal(data, &m); err == nil {
			if m.Name != "" {
				name = m.Name
			}
			if m.Version != "" {
				localVersion = m.Version
			}
			repoURL = m.Repository
		}
	}

	gitInfo := GetGitInfo(ctx, dirPath)
	isSymlink, _ := fsutil.IsSymlink(dirPath)
	isGit := gitInfo != nil

	remoteVersion := "n/a"
	status := "UP-TO-DATE"
	details := ""

	owner, repo, repoParsed := ParseRepoOwnerName(repoURL)

	if isGit && gitInfo.RemoteURL != "" {
		remoteHead := GetRemoteGitHead(ctx, gitInfo.RemoteURL)
		localCommit := gitInfo.Commit
		if len(localCommit) > 7 {
			localCommit = localCommit[:7]
		}

		if remoteHead != "" {
			remoteShort := remoteHead
			if len(remoteShort) > 7 {
				remoteShort = remoteShort[:7]
			}
			remoteVersion = remoteShort
			if localCommit != remoteShort {
				status = "OUTDATED"
				details = "commit " + localCommit + " -> " + remoteShort
			} else {
				status = "UP-TO-DATE"
				details = "git @ " + localCommit
			}
		} else {
			remoteVersion = localCommit
			details = "local git @ " + localCommit
		}
	} else if repoParsed {
		remoteInfo := GetRemoteGitHubLatest(ctx, owner, repo)
		if remoteInfo != nil {
			remoteVersion = remoteInfo.Latest
			cleanLocal := strings.TrimPrefix(localVersion, "v")
			cleanRemote := strings.TrimPrefix(remoteVersion, "v")

			if cleanLocal != "unknown" && cleanLocal != cleanRemote {
				status = "OUTDATED"
				details = cleanLocal + " -> " + cleanRemote
			} else {
				status = "UP-TO-DATE"
				details = remoteInfo.Type + ": " + cleanRemote
			}
		}
	}

	return AuditResult{
		Name:          name,
		Path:          dirPath,
		IsSymlink:     isSymlink,
		IsGit:         isGit,
		LocalVersion:  localVersion,
		RemoteVersion: remoteVersion,
		Status:        status,
		RepoURL:       repoURL,
		Details:       details,
	}
}
