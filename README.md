# 🎛️ Antigravity Hub (`antigravity-hub`)

> A modular persona switcher, minimalist skill & subagent curator, and plugin version audit engine for **Google Antigravity (`agy` CLI and Hub)**.

Antigravity Hub solves **customization bloat** by keeping your active context clean and minimalist, allowing you to load and unload skills, subagents, and plugins on-demand based on your active persona or immediate task. It also provides an automated version-checking engine that queries upstream Git repositories and GitHub releases to keep your environment up to date.

---

## 🎯 Key Features

1. **Minimalist by Default**:
   - Reduces active skills in everyday sessions from **60+ to < 5**, cutting context overhead and accelerating inference.
   - Moves bulk and inactive skills into an offline central catalog (`~/.gemini/catalog/skills/`).

2. **Persona & Task Profiles**:
   - Pre-configured profiles that toggle corresponding plugins in `config.json` and symlink only relevant skills:
     - `core`: Minimalist baseline (< 5 skills) for everyday development, dotfiles, and general coding.
     - `architect`: Bean-to-Cup SDLC Stage 0-4 (`ideator`, `grill`, `write-prd`, `red-team-reviewer`).
     - `fullstack`: Modern web development, React/Vite, UI components, Chrome DevTools debugging.
     - `stitch`: Google Stitch design system, screen generation, and Remotion video walkthroughs.
     - `gcp-sre`: Google Cloud infrastructure, Cloud Run, GKE, BigQuery, AlloyDB, and chaos mitigation.
     - `adk-dev`: Agent development with Google ADK, Agent Platform, and Antigravity SDK.

3. **Dynamic Load & Unload**:
   - Switch personas instantly with a single command (`agy-persona switch stitch`).
   - Temporarily load or unload individual skills or plugins on the fly (`agy-persona load gcloud`).
   - Reset back to the minimal baseline anytime (`agy-persona reset`).

4. **Plugin Version & Update Engine**:
   - Audits all installed plugins in `~/.gemini/config/plugins/` against upstream GitHub repositories and Git remotes.
   - Compares local versions and Git commit SHAs against the latest remote releases and HEAD commits.
   - One-step automated update command: `agy-plugins update [name|all]`.

5. **Non-Blocking Launch Hook**:
   - Integrates with Antigravity via `hooks.json` (`SessionStart`).
   - Uses a cached background check (TTL: 6 hours) so CLI launch takes **< 20ms**.
   - If updates are detected, injects a clean Markdown notification directly into your session context upon launch.

---

## 🚀 Quick Start & Installation

Run the automated setup script to catalog existing skills, install CLI binaries, and register the launch hook:

```bash
cd ~/workspace/antigravity-hub
./scripts/setup.sh
```

This will:
1. Populate `~/.gemini/catalog/skills/` with existing skills.
2. Link `agy-hub`, `agy-persona`, and `agy-plugins` into `~/.local/bin/`.
3. Register `antigravity-hub` as an Antigravity plugin in `~/.gemini/config/plugins/`.
4. Apply the `core` minimalist persona.

---

## 📖 CLI Usage

### 1. Persona Management (`agy-persona`)

```bash
# List all available personas and active status
agy-persona list

# Switch to a persona
agy-persona switch stitch
agy-persona switch architect
agy-persona switch fullstack
agy-persona switch gcp-sre
agy-persona switch core

# Show active persona, loaded skills, and enabled plugins
agy-persona current

# Temporarily load or unload an individual component
agy-persona load gcloud
agy-persona unload gcloud

# Reset back to minimal core
agy-persona reset
```

### 2. Plugin Version Audit & Updates (`agy-plugins`)

```bash
# Display audit table of all installed plugins vs upstream
agy-plugins check

# View updates formatted as markdown
agy-plugins check --markdown

# Output as JSON
agy-plugins check --json

# Update all outdated plugins
agy-plugins update all

# Update a specific plugin
agy-plugins update bean-to-cup
```

### 3. Unified Dispatcher (`agy-hub`)

```bash
agy-hub persona switch stitch
agy-hub plugins check
agy-hub plugins update
```

---

## 🛠️ Architecture & Discovery Alignment

Antigravity loads customizations according to specific priority rules:
1. **Workspace Project**: Hierarchical discovery from CWD to repo root (`.agents/`).
2. **Global Configuration**: Discovered under `~/.gemini/config/skills/` and `~/.gemini/config/plugins/`.
3. **Plugin Activation**: Toggled via `~/.gemini/config/config.json` (`plugins.<name>.enabled = true/false`).

Antigravity Hub respects these native primitives without modifying core Antigravity binaries:
- Personas update `config.json` plugin flags natively.
- Active skills in `~/.gemini/config/skills/` are symlinked directly from `~/.gemini/catalog/skills/`.
- Unmanaged workspace sprawl is quarantined so every session starts lightning-fast.

---

## 🧪 Testing

Run the test suite:
```bash
python3 -m unittest discover tests/
```

---

## 📄 License

Apache-2.0
