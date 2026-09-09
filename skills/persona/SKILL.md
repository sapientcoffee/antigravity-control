---
name: persona
description: >-
  Antigravity Persona Switcher. Use this skill when the user asks to switch personas,
  load or unload skills/plugins, or reset to a minimalist core configuration.
---

# Antigravity Persona Switcher

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
agy-persona switch <persona-name>
```

### Show Current Persona
```bash
agy-persona current
```

### List Available Personas
```bash
agy-persona list
```

### Temporarily Load / Unload Component
```bash
agy-persona load <skill-or-plugin-name>
agy-persona unload <skill-or-plugin-name>
```

### Reset to Minimal Core
```bash
agy-persona reset
```
