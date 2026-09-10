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
agyctl plugins check
```
Or view formatted markdown:
```bash
agyctl plugins check --markdown
```

### 2. Update Outdated Plugins
Update all outdated plugins:
```bash
agyctl plugins update all
```
Or update a specific plugin:
```bash
agyctl plugins update bean-to-cup
```

### 3. List Installed Plugins
```bash
agyctl plugins list
```
