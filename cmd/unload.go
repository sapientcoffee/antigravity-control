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

	"github.com/sapientcoffee/antigravity-control/pkg/config"
	"github.com/sapientcoffee/antigravity-control/pkg/persona"
	"github.com/spf13/cobra"
)

// NewUnloadCmd creates the command to unload an individual skill or plugin.
func NewUnloadCmd(pathsFn func() (*config.Paths, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unload <name>",
		Short: "Temporarily unload a skill or plugin",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			paths, err := pathsFn()
			if err != nil {
				return err
			}

			itemName := args[0]
			itemType, err := persona.UnloadItem(paths, itemName)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			if itemType == persona.ItemPlugin {
				fmt.Fprintf(out, " Plugin disabled: \033[1;33m%s\033[0m\n", itemName)
			} else {
				fmt.Fprintf(out, " Skill unloaded: \033[1;33m%s\033[0m\n", itemName)
			}

			return nil
		},
	}

	return cmd
}
