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
	"github.com/sapientcoffee/antigravity-control/pkg/config"
	"github.com/spf13/cobra"
)

// NewPersonaCmd creates the grouped 'persona' command.
func NewPersonaCmd(pathsFn func() (*config.Paths, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "persona",
		Short: "Manage Antigravity persona profiles",
		Long:  "Grouped command for inspecting, listing, switching, and resetting persona profiles.",
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

	return cmd
}
