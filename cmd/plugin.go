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
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/sapientcoffee/antigravity-control/pkg/config"
	"github.com/sapientcoffee/antigravity-control/pkg/plugin"
	"github.com/spf13/cobra"
)

// NewPluginCmd creates the grouped 'plugin' / 'plugins' command.
func NewPluginCmd(pathsFn func() (*config.Paths, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "plugin",
		Aliases: []string{"plugins"},
		Short:   "Manage and audit Antigravity plugins",
		Long:    "Inspect, audit versions, and update Antigravity plugins against their upstream Git remotes and GitHub releases.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Default action when run without subcommand: list plugins
			listCmd := NewPluginListCmd(pathsFn)
			return listCmd.RunE(cmd, args)
		},
	}

	cmd.AddCommand(NewPluginListCmd(pathsFn))
	cmd.AddCommand(NewCheckCmd(pathsFn))
	cmd.AddCommand(NewUpdateCmd(pathsFn))

	return cmd
}

// NewPluginListCmd creates the subcommand to list installed plugins.
func NewPluginListCmd(pathsFn func() (*config.Paths, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List installed plugin manifests and local versions",
		RunE: func(cmd *cobra.Command, args []string) error {
			paths, err := pathsFn()
			if err != nil {
				return err
			}

			entries, err := os.ReadDir(paths.PluginsDir)
			if err != nil && !os.IsNotExist(err) {
				return err
			}

			type pluginSummary struct {
				Name        string
				Version     string
				Description string
			}
			var plugins []pluginSummary

			for _, entry := range entries {
				dirPath := filepath.Join(paths.PluginsDir, entry.Name())
				fi, err := os.Stat(dirPath)
				if err != nil || !fi.IsDir() {
					continue
				}

				name := entry.Name()
				version := "unknown"
				description := ""

				manifestFile := filepath.Join(dirPath, "plugin.json")
				if data, err := os.ReadFile(manifestFile); err == nil {
					var m plugin.Manifest
					if err := json.Unmarshal(data, &m); err == nil {
						if m.Name != "" {
							name = m.Name
						}
						if m.Version != "" {
							version = m.Version
						}
						description = m.Description
					}
				}

				plugins = append(plugins, pluginSummary{
					Name:        name,
					Version:     version,
					Description: description,
				})
			}

			sort.Slice(plugins, func(i, j int) bool {
				return plugins[i].Name < plugins[j].Name
			})

			out := cmd.OutOrStdout()
			fmt.Fprintln(out, "\n\033[1;35mInstalled Antigravity Plugins\033[0m")
			fmt.Fprintln(out, strings.Repeat("=", 70))
			fmt.Fprintf(out, "%-30s %-12s %s\n", "PLUGIN", "VERSION", "DESCRIPTION")
			fmt.Fprintln(out, strings.Repeat("-", 70))

			for _, p := range plugins {
				desc := p.Description
				if len(desc) > 35 {
					desc = desc[:32] + "..."
				}
				fmt.Fprintf(out, "%-30s %-12s %s\n", p.Name, p.Version, desc)
			}
			fmt.Fprintln(out, strings.Repeat("-", 70))

			return nil
		},
	}

	return cmd
}
