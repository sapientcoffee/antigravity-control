---
name: persona
description: >-
  Antigravity Persona Switcher. Use this skill when the user asks to switch personas,
  load or unload skills/plugins, or reset to a minimalist core configuration using agyctl.
---

# Antigravity Persona Switcher (`agyctl persona`)

Use this skill to dynamically reconfigure Antigravity's active plugins and skills based on the desired persona or task.

## Persona Profiles

- **`core`**: Minimalist baseline (< 5 skills). Everyday coding, maintenance, and dotfiles.
- **`architect`**: Bean-to-Cup SDLC Stage 0-4 (ideation, socratic interview, PRD writing, red-team auditing).
- **`fullstack`**: Modern web development, React/Vite, UI components, Chrome DevTools debugging.
- **`stitch`**: Google Stitch design systems, UI generation, and Remotion video walkthroughs.
- **`gcp-sre`**: Google Cloud Platform infrastructure, databases, networking, and observability.
- **`adk-dev`**: AI Agent development with Google ADK, Agent Platform, and Antigravity SDK.

## Commands

### Switch Persona
```bash
agyctl persona switch <persona-name>
# or shortcut:
agyctl switch <persona-name>
```

### Show Current Persona & Config Location
```bash
agyctl persona current
# or shortcut:
agyctl current
```

### Inspect Persona Config File Location
```bash
# Print the JSON config file path of the active persona:
agyctl persona path
# or shortcut:
agyctl path

# Print the JSON config file path of a specific persona:
agyctl persona path <persona-name>

# Print the personas configuration directory:
agyctl persona path --dir
```

### List Available Personas
```bash
agyctl persona list
# or shortcut:
agyctl personas
```

### Temporarily Load / Unload Component
```bash
agyctl load <skill-or-plugin-name>
agyctl unload <skill-or-plugin-name>
```

### Reset to Minimal Core
```bash
agyctl reset
```
