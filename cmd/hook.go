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
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sapientcoffee/antigravity-control/pkg/config"
	"github.com/sapientcoffee/antigravity-control/pkg/fsutil"
	"github.com/sapientcoffee/antigravity-control/pkg/plugin"
	"github.com/spf13/cobra"
)

const hookCheckTTL = 6 * time.Hour

// NewHookCmd creates the command group for Antigravity lifecycle hooks.
func NewHookCmd(pathsFn func() (*config.Paths, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hook",
		Short: "Antigravity lifecycle hook integration",
	}

	cmd.AddCommand(NewHookSessionStartCmd(pathsFn))
	return cmd
}

// NewHookSessionStartCmd handles the SessionStart lifecycle hook.
// It checks the 6-hour cache TTL and injects upstream update notices if plugins are outdated.
func NewHookSessionStartCmd(pathsFn func() (*config.Paths, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "session-start",
		Short: "Execute SessionStart freshness check (cached, non-blocking)",
		RunE: func(cmd *cobra.Command, args []string) error {
			paths, err := pathsFn()
			if err != nil {
				return err
			}

			now := time.Now()
			lastCheckTime := time.Time{}

			if tsData, err := os.ReadFile(paths.HookTimestampFile); err == nil {
				if sec, err := strconv.ParseInt(strings.TrimSpace(string(tsData)), 10, 64); err == nil {
					lastCheckTime = time.Unix(sec, 0)
				}
			}

			noticeExists := false
			if fi, err := os.Stat(paths.HookCachedNoticeFile); err == nil && !fi.IsDir() {
				noticeExists = true
			}

			needsAudit := !noticeExists || now.Sub(lastCheckTime) > hookCheckTTL

			if needsAudit {
				results, err := plugin.AuditPlugins(cmd.Context(), paths)
				if err == nil {
					var outdated []plugin.AuditResult
					for _, r := range results {
						if r.Status == "OUTDATED" {
							outdated = append(outdated, r)
						}
					}

					var buf bytes.Buffer
					if len(outdated) > 0 {
						RenderMarkdownNotice(&buf, outdated)
					}
					_ = fsutil.AtomicWriteFile(paths.HookCachedNoticeFile, buf.Bytes(), 0644)
					tsStr := fmt.Sprintf("%d\n", now.Unix())
					_ = fsutil.AtomicWriteFile(paths.HookTimestampFile, []byte(tsStr), 0644)
				}
			}

			// If cached notice has content, print to stdout for Antigravity session injection
			if noticeData, err := os.ReadFile(paths.HookCachedNoticeFile); err == nil && len(noticeData) > 0 {
				_, _ = cmd.OutOrStdout().Write(noticeData)
			}

			return nil
		},
	}

	return cmd
}
