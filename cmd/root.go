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
	"github.com/spf13/cobra"
)

var geminiDirOverride string

// Execute executes the root CLI command.
func Execute() error {
	rootCmd := NewRootCmd()
	return rootCmd.Execute()
}

// NewRootCmd creates the top-level agyctl command and registers all subcommands.
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "agyctl",
		Short: "Antigravity Control (agyctl) - Dynamic Persona and Plugin Manager",
		Long: `agyctl (Antigravity Control) manages active personas by updating Antigravity's
plugin toggles and symlinking only relevant skills into the active skills discovery directory,
preventing context bloat and keeping your environment minimalist and responsive.

Persona configuration JSON files are located in ~/.gemini/personas/ (e.g. core.json).`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	rootCmd.PersistentFlags().StringVar(&geminiDirOverride, "gemini-dir", "", "override base Antigravity directory (default ~/.gemini)")

	pathsFn := func() (*config.Paths, error) {
		return config.ResolvePaths(geminiDirOverride)
	}

	// Persona commands (direct shortcuts & persona group)
	rootCmd.AddCommand(NewSwitchCmd(pathsFn))
	rootCmd.AddCommand(NewCurrentCmd(pathsFn))
	rootCmd.AddCommand(NewListCmd(pathsFn))
	rootCmd.AddCommand(NewResetCmd(pathsFn))
	rootCmd.AddCommand(NewLoadCmd(pathsFn))
	rootCmd.AddCommand(NewUnloadCmd(pathsFn))
	rootCmd.AddCommand(NewPersonaPathCmd(pathsFn))
	rootCmd.AddCommand(NewPersonaCmd(pathsFn))

	// Plugin commands (direct shortcuts & plugin group)
	rootCmd.AddCommand(NewCheckCmd(pathsFn))
	rootCmd.AddCommand(NewUpdateCmd(pathsFn))
	rootCmd.AddCommand(NewPluginCmd(pathsFn))

	// Lifecycle hooks
	rootCmd.AddCommand(NewHookCmd(pathsFn))

	// Version
	rootCmd.AddCommand(NewVersionCmd())

	return rootCmd
}

// CheckErr prints any command failure and exits.
func CheckErr(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
