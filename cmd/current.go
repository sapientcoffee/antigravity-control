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
	"fmt"
	"path/filepath"
	"strings"

	"github.com/sapientcoffee/antigravity-control/pkg/config"
	"github.com/sapientcoffee/antigravity-control/pkg/persona"
	"github.com/spf13/cobra"
)

// NewCurrentCmd creates the command to show the active persona and components.
func NewCurrentCmd(pathsFn func() (*config.Paths, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "current",
		Aliases: []string{"status"},
		Short:   "Show active persona, loaded skills, and enabled plugins",
		RunE: func(cmd *cobra.Command, args []string) error {
			paths, err := pathsFn()
			if err != nil {
				return err
			}

			status, err := persona.GetCurrent(paths)
			if err != nil {
				return err
			}

			displayName := status.ActivePersonaName
			description := "N/A"
			configLocation := "N/A"
			if status.Persona != nil {
				if status.Persona.DisplayName != "" {
					displayName = status.Persona.DisplayName
				}
				if status.Persona.Description != "" {
					description = status.Persona.Description
				}
				if status.Persona.ConfigPath != "" {
					configLocation = status.Persona.ConfigPath
				}
			} else if paths.PersonasDir != "" {
				configLocation = filepath.Join(paths.PersonasDir, status.ActivePersonaName+".json") + " (not found)"
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Active Persona : \033[1;36m%s\033[0m (%s)\n", displayName, status.ActivePersonaName)
			fmt.Fprintf(out, "Config Location: %s\n", configLocation)
			fmt.Fprintf(out, "Description    : %s\n", description)

			skillsSample := status.ActiveSkills
			suffix := ""
			if len(skillsSample) > 8 {
				skillsSample = skillsSample[:8]
				suffix = "..."
			}
			fmt.Fprintf(out, "Active Skills  (%d): %s%s\n", len(status.ActiveSkills), strings.Join(skillsSample, ", "), suffix)
			fmt.Fprintf(out, "Active Plugins (%d): %s\n", len(status.ActivePlugins), strings.Join(status.ActivePlugins, ", "))

			return nil
		},
	}

	return cmd
}
