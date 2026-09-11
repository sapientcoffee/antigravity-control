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
	"strings"

	"github.com/sapientcoffee/antigravity-control/pkg/config"
	"github.com/sapientcoffee/antigravity-control/pkg/persona"
	"github.com/spf13/cobra"
)

// NewSwitchCmd creates the command to switch the active persona profile.
func NewSwitchCmd(pathsFn func() (*config.Paths, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "switch <persona>",
		Short: "Switch active persona profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			paths, err := pathsFn()
			if err != nil {
				return err
			}

			targetPersona := args[0]
			result, err := persona.Switch(paths, targetPersona)
			if err != nil {
				return err
			}

			dispName := result.Persona.DisplayName
			if dispName == "" {
				dispName = result.Persona.Name
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, " switched to persona: \033[1;36m%s\033[0m (%s)\n", dispName, targetPersona)
			if result.Persona.ConfigPath != "" {
				fmt.Fprintf(out, "   Config Location : %s\n", result.Persona.ConfigPath)
			}
			if result.Persona.Description != "" {
				fmt.Fprintf(out, "   Description     : %s\n", result.Persona.Description)
			}
			fmt.Fprintf(out, "   Plugins enabled : %s\n", strings.Join(result.EnabledPlugins, ", "))
			fmt.Fprintf(out, "   Skills active   : %d loaded\n", result.LinkedCount)
			if len(result.MissingSkills) > 0 {
				sample := result.MissingSkills
				if len(sample) > 3 {
					sample = sample[:3]
				}
				fmt.Fprintf(out, "   Notice: %d skills not found in catalog (%s...)\n", len(result.MissingSkills), strings.Join(sample, ", "))
			}

			return nil
		},
	}

	return cmd
}
