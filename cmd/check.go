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
	"io"
	"os"
	"strings"

	"github.com/sapientcoffee/antigravity-control/pkg/config"
	"github.com/sapientcoffee/antigravity-control/pkg/plugin"
	"github.com/spf13/cobra"
)

// NewCheckCmd creates the command to check plugins against upstream repositories.
func NewCheckCmd(pathsFn func() (*config.Paths, error)) *cobra.Command {
	var jsonOutput bool
	var markdownOutput bool
	var quietOutput bool

	cmd := &cobra.Command{
		Use:   "check",
		Short: "Audit installed plugins against upstream GitHub & Git remotes",
		RunE: func(cmd *cobra.Command, args []string) error {
			paths, err := pathsFn()
			if err != nil {
				return err
			}

			results, err := plugin.AuditPlugins(cmd.Context(), paths)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			var outdated []plugin.AuditResult
			for _, r := range results {
				if r.Status == "OUTDATED" {
					outdated = append(outdated, r)
				}
			}

			if jsonOutput {
				enc := json.NewEncoder(out)
				enc.SetIndent("", "  ")
				return enc.Encode(results)
			}

			if markdownOutput {
				RenderMarkdownNotice(out, outdated)
				if len(outdated) > 0 {
					os.Exit(1)
				}
				return nil
			}

			if !quietOutput {
				RenderAuditTable(out, results)
			}

			if len(outdated) > 0 {
				os.Exit(1)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	cmd.Flags().BoolVar(&markdownOutput, "markdown", false, "Output as Markdown (for session hooks)")
	cmd.Flags().BoolVar(&quietOutput, "quiet", false, "Quiet output, exit code 1 if outdated")

	return cmd
}

// RenderAuditTable prints the formatted console table of plugin audit results.
func RenderAuditTable(out io.Writer, results []plugin.AuditResult) {
	fmt.Fprintln(out, "\n\033[1;35mAntigravity Plugin Version Audit\033[0m")
	fmt.Fprintln(out, strings.Repeat("=", 86))
	fmt.Fprintf(out, "%-28s %-14s %-14s %-12s %-14s\n", "PLUGIN", "LOCAL VER", "REMOTE", "STATUS", "DETAILS")
	fmt.Fprintln(out, strings.Repeat("-", 86))

	for _, item := range results {
		statusColor := "\033[1;32m"
		if item.Status != "UP-TO-DATE" {
			statusColor = "\033[1;31m"
		}
		fmt.Fprintf(
			out,
			"%-28s %-14s %-14s %s%-12s\033[0m %-14s\n",
			item.Name,
			item.LocalVersion,
			item.RemoteVersion,
			statusColor,
			item.Status,
			item.Details,
		)
	}
	fmt.Fprintln(out, strings.Repeat("-", 86))

	var outdated []plugin.AuditResult
	for _, r := range results {
		if r.Status == "OUTDATED" {
			outdated = append(outdated, r)
		}
	}

	if len(outdated) > 0 {
		fmt.Fprintf(out, "\033[1;33m💡 %d plugin(s) have updates available!\033[0m Run \033[1;36magyctl update [name|all]\033[0m to update.\n\n", len(outdated))
	} else {
		fmt.Fprintln(out, "\033[1;32m All Antigravity plugins are up-to-date.\033[0m")
	}
}

// RenderMarkdownNotice formats outdated plugins into a session notice.
func RenderMarkdownNotice(out io.Writer, outdated []plugin.AuditResult) {
	if len(outdated) == 0 {
		return
	}
	fmt.Fprintln(out, "### 🔄 Antigravity Plugin Updates Available")
	fmt.Fprintf(out, "> Found **%d** plugin(s) with newer versions upstream:\n\n", len(outdated))
	for _, item := range outdated {
		details := ""
		if item.Details != "" {
			details = " (" + item.Details + ")"
		}
		fmt.Fprintf(out, "- **`%s`**: `%s` ➔ `%s`%s\n", item.Name, item.LocalVersion, item.RemoteVersion, details)
	}
	fmt.Fprintln(out, "\nRun `agyctl update` to update.")
}
