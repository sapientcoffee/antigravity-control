#!/bin/bash
# Copyright 2026 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

HOME_DIR="$HOME"
CONFIG_DIR="$HOME_DIR/.gemini/config"
PLUGINS_DIR="$CONFIG_DIR/plugins"
SKILLS_DIR="$CONFIG_DIR/skills"
CATALOG_DIR="$HOME_DIR/.gemini/catalog/skills"
BIN_DIR="$HOME_DIR/.local/bin"
AGENTS_SKILLS_DIR="$HOME_DIR/.agents/skills"

echo -e "\033[1;35m==> Installing Antigravity Control (agyctl)...\033[0m"

# 1. Ensure catalog directory exists
mkdir -p "$CATALOG_DIR"
mkdir -p "$BIN_DIR"
mkdir -p "$PLUGINS_DIR"
mkdir -p "$SKILLS_DIR"

# 2. Populate Catalog from ~/.agents/skills (if present)
if [ -d "$AGENTS_SKILLS_DIR" ]; then
    echo "Cataloging existing skills from $AGENTS_SKILLS_DIR to $CATALOG_DIR..."
    for skill_path in "$AGENTS_SKILLS_DIR"/*; do
        if [ -d "$skill_path" ]; then
            skill_name="$(basename "$skill_path")"
            if [ ! -e "$CATALOG_DIR/$skill_name" ]; then
                cp -r "$skill_path" "$CATALOG_DIR/"
            fi
        fi
    done
fi

# Also catalog any skills found in plugins
for plugin_skill_dir in "$PLUGINS_DIR"/*/skills/*; do
    if [ -d "$plugin_skill_dir" ]; then
        skill_name="$(basename "$plugin_skill_dir")"
        if [ ! -e "$CATALOG_DIR/$skill_name" ]; then
            cp -r "$plugin_skill_dir" "$CATALOG_DIR/"
        fi
    fi
done

echo " Catalog populated with $(ls -1 "$CATALOG_DIR" | wc -l) skills."

# 3. Install CLI binaries to ~/.local/bin
echo "Installing agyctl to $BIN_DIR..."
ln -sf "$REPO_DIR/bin/agyctl" "$BIN_DIR/agyctl"

# Remove deprecated hub links if present
rm -f "$BIN_DIR/agy-hub" "$BIN_DIR/agy-persona" "$BIN_DIR/agy-plugins"
rm -f "$PLUGINS_DIR/antigravity-hub"

# 4. Link plugin to Antigravity plugins directory so hooks and skills are active
echo "Registering antigravity-control plugin with Antigravity..."
ln -sfn "$REPO_DIR" "$PLUGINS_DIR/antigravity-control"

# 5. Clean up old bloated symlinks in ~/.agents/skills so $HOME is no longer cluttered
if [ -d "$AGENTS_SKILLS_DIR" ]; then
    # Only backup if non-empty
    if [ "$(ls -A "$AGENTS_SKILLS_DIR" 2>/dev/null)" ]; then
        BACKUP_DIR="$HOME_DIR/.agents/skills_backup_$(date +%Y%m%d_%H%M%S)"
        echo "Backing up and resetting $AGENTS_SKILLS_DIR to $BACKUP_DIR..."
        mv "$AGENTS_SKILLS_DIR" "$BACKUP_DIR"
        mkdir -p "$AGENTS_SKILLS_DIR"
    fi
fi

# 6. Apply minimal core persona
echo "Applying default 'core' minimalist persona..."
python3 "$REPO_DIR/scripts/persona_manager.py" switch core

echo ""
echo -e "\033[1;32m🎉 Antigravity Control (agyctl) installed successfully!\033[0m"
echo "Run 'agyctl list' to inspect available personas."
echo "Run 'agyctl check' to audit plugin versions."
