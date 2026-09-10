---
name: control
description: >-
  Antigravity Control orchestrator. Use this skill when the user asks to manage, audit,
  check updates for, or update Antigravity plugins, skills, or personas via agyctl.
---

# Antigravity Control Skill (`agyctl`)

Use this skill to inspect and update Antigravity plugins and manage their upstream lifecycles.

## Available Actions

### 1. Check Plugin Versions & Upstream Updates
Run the plugin checker:
```bash
agyctl check
# or grouped command:
agyctl plugins check
```
View formatted markdown (used for session hooks):
```bash
agyctl check --markdown
```
Or output as JSON:
```bash
agyctl check --json
```

### 2. Update Outdated Plugins
Update all outdated plugins:
```bash
agyctl update all
# or grouped command:
agyctl plugins update all
```
Or update a specific plugin:
```bash
agyctl update <plugin-name>
```

### 3. List Installed Plugins
```bash
agyctl plugins
# or:
agyctl plugins list
```

### 4. Lifecycle Session Hook
Execute the non-blocking upstream freshness check (cached with 6-hour TTL):
```bash
agyctl hook session-start
```
