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

// Persona represents an Antigravity persona configuration profile.
type Persona struct {
	Name            string   `json:"name"`
	DisplayName     string   `json:"displayName"`
	Description     string   `json:"description"`
	EnabledPlugins  []string `json:"enabled_plugins"`
	DisabledPlugins []string `json:"disabled_plugins"`
	ActiveSkills    []string `json:"active_skills"`
}

// SwitchResult contains the outcome of a persona switch operation.
type SwitchResult struct {
	Persona        Persona
	EnabledPlugins []string
	LinkedCount    int
	MissingSkills  []string
}

// CurrentStatus represents the live active persona, loaded skills, and enabled plugins.
type CurrentStatus struct {
	ActivePersonaName string
	Persona           *Persona
	ActiveSkills      []string
	ActivePlugins     []string
}

// ItemType indicates whether a loaded/unloaded entity is a plugin or a skill.
type ItemType string

const (
	ItemPlugin ItemType = "plugin"
	ItemSkill  ItemType = "skill"
)
