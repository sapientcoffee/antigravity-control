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
	"os"
	"path/filepath"

	"github.com/sapientcoffee/antigravity-control/pkg/config"
	"github.com/sapientcoffee/antigravity-control/pkg/persona"
	"github.com/spf13/cobra"
)

// NewPersonaCmd creates the grouped 'persona' command.
func NewPersonaCmd(pathsFn func() (*config.Paths, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "persona",
		Short: "Manage Antigravity persona profiles",
		Long: `Grouped command for inspecting, listing, switching, and resetting persona profiles.
Persona definitions are stored as JSON files (default: ~/.gemini/personas/<name>.json).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Default to showing current persona when run without subcommand
			currentCmd := NewCurrentCmd(pathsFn)
			return currentCmd.RunE(cmd, args)
		},
	}

	cmd.AddCommand(NewSwitchCmd(pathsFn))
	cmd.AddCommand(NewCurrentCmd(pathsFn))
	cmd.AddCommand(NewListCmd(pathsFn))
	cmd.AddCommand(NewResetCmd(pathsFn))
	cmd.AddCommand(NewPersonaPathCmd(pathsFn))

	return cmd
}

// NewPersonaPathCmd creates the command to show the filesystem path of a persona's JSON config.
func NewPersonaPathCmd(pathsFn func() (*config.Paths, error)) *cobra.Command {
	var showDir bool
	cmd := &cobra.Command{
		Use:     "path [persona]",
		Aliases: []string{"config", "file"},
		Short:   "Print the filesystem path of the persona JSON config",
		Long: `Print the absolute filesystem location of a persona's JSON configuration file.
If no persona name is provided, prints the location of the currently active persona.
Use --dir to print the directory containing persona definitions.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			paths, err := pathsFn()
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			if showDir {
				fmt.Fprintln(out, paths.PersonasDir)
				return nil
			}

			target := ""
			if len(args) > 0 {
				target = args[0]
			} else {
				target = config.GetActivePersona(paths.StateFile)
			}

			personas, err := persona.GetAvailablePersonas(paths)
			if err != nil {
				return err
			}

			if p, ok := personas[target]; ok && p.ConfigPath != "" {
				fmt.Fprintln(out, p.ConfigPath)
				return nil
			}

			// Check if candidate file exists in PersonasDir
			if paths.PersonasDir != "" {
				candidate := filepath.Join(paths.PersonasDir, target+".json")
				if _, err := os.Stat(candidate); err == nil {
					fmt.Fprintln(out, candidate)
					return nil
				}
			}

			return fmt.Errorf("persona config for %q not found in %s", target, paths.PersonasDir)
		},
	}

	cmd.Flags().BoolVarP(&showDir, "dir", "d", false, "print the persona configurations directory instead of file path")
	return cmd
}
