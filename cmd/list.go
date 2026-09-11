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
	"sort"
	"strings"

	"github.com/sapientcoffee/antigravity-control/pkg/config"
	"github.com/sapientcoffee/antigravity-control/pkg/persona"
	"github.com/spf13/cobra"
)

// NewListCmd creates the command to list all available persona profiles.
func NewListCmd(pathsFn func() (*config.Paths, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"personas"},
		Short:   "List all available persona profiles",
		RunE: func(cmd *cobra.Command, args []string) error {
			paths, err := pathsFn()
			if err != nil {
				return err
			}

			personas, err := persona.GetAvailablePersonas(paths)
			if err != nil {
				return err
			}

			current := config.GetActivePersona(paths.StateFile)

			out := cmd.OutOrStdout()
			dirInfo := ""
			if paths.PersonasDir != "" {
				dirInfo = fmt.Sprintf(" (%s)", paths.PersonasDir)
			}
			fmt.Fprintf(out, "\n\033[1;35mAntigravity Personas\033[0m%s\n", dirInfo)
			fmt.Fprintln(out, strings.Repeat("=", 60))

			var names []string
			for name := range personas {
				names = append(names, name)
			}
			sort.Strings(names)

			for _, name := range names {
				data := personas[name]
				marker := "  "
				if name == current {
					marker = "\033[1;32m* (active)\033[0m"
				}

				dispName := data.DisplayName
				if dispName == "" {
					dispName = name
				}

				fmt.Fprintf(out, "%s \033[1;37m%-14s\033[0m : %s\n", marker, name, dispName)
				if data.Description != "" {
					fmt.Fprintf(out, "     %s\n", data.Description)
				}
				if data.ConfigPath != "" {
					fmt.Fprintf(out, "     Config  : %s\n", data.ConfigPath)
				}
				fmt.Fprintf(out, "     Plugins : %s\n", strings.Join(data.EnabledPlugins, ", "))
				fmt.Fprintf(out, "     Skills  : %d configured\n\n", len(data.ActiveSkills))
			}

			return nil
		},
	}

	return cmd
}
