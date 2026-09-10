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

// AuditResult represents the audit status of an installed plugin.
type AuditResult struct {
	Name          string `json:"name"`
	Path          string `json:"path"`
	IsSymlink     bool   `json:"is_symlink"`
	IsGit         bool   `json:"is_git"`
	LocalVersion  string `json:"local_version"`
	RemoteVersion string `json:"remote_version"`
	Status        string `json:"status"` // "UP-TO-DATE" or "OUTDATED"
	RepoURL       string `json:"repo_url"`
	Details       string `json:"details"`
}

// Manifest represents the structure of plugin.json.
type Manifest struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Repository  string `json:"repository"`
	Description string `json:"description"`
}
