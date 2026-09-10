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

	"github.com/sapientcoffee/antigravity-control/pkg/config"
	"github.com/sapientcoffee/antigravity-control/pkg/plugin"
	"github.com/spf13/cobra"
)

// NewUpdateCmd creates the command to update outdated plugins.
func NewUpdateCmd(pathsFn func() (*config.Paths, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update [name|all]",
		Short: "Update outdated plugins to latest upstream releases",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			paths, err := pathsFn()
			if err != nil {
				return err
			}

			target := "all"
			if len(args) > 0 && args[0] != "" {
				target = args[0]
			}

			out := cmd.OutOrStdout()
			fmt.Fprintln(out, "Auditing plugins for available updates...")
			results, err := plugin.AuditPlugins(cmd.Context(), paths)
			if err != nil {
				return err
			}

			var toUpdate []plugin.AuditResult
			for _, r := range results {
				if (target == "all" || r.Name == target) && r.Status == "OUTDATED" {
					toUpdate = append(toUpdate, r)
				}
			}

			if len(toUpdate) == 0 {
				fmt.Fprintf(out, "No outdated plugins found matching '%s'.\n", target)
				return nil
			}

			success := 0
			for _, item := range toUpdate {
				fmt.Fprintf(out, "Updating plugin: \033[1;36m%s\033[0m...\n", item.Name)
				if err := plugin.UpdatePlugin(cmd.Context(), item); err != nil {
					fmt.Fprintf(os.Stderr, "  Error updating %s: %v\n", item.Name, err)
				} else {
					fmt.Fprintf(out, "  Successfully updated %s.\n", item.Name)
					success++
				}
			}

			fmt.Fprintf(out, "\nUpdated %d/%d plugin(s).\n", success, len(toUpdate))
			return nil
		},
	}

	return cmd
}
