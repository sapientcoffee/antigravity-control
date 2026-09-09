#!/usr/bin/env python3
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

"""Antigravity Persona and Skill/Plugin Switcher.

Manages active personas by updating Antigravity's config.json plugin toggles
and symlinking only relevant skills from the central catalog into the active
skills discovery directory.
"""

import argparse
import json
import os
from pathlib import Path
import shutil
import sys

HOME = Path.home()
CONFIG_DIR = HOME / ".gemini" / "config"
CONFIG_FILE = CONFIG_DIR / "config.json"
SKILLS_DIR = CONFIG_DIR / "skills"
CATALOG_DIR = HOME / ".gemini" / "catalog" / "skills"
PERSONAS_DIR = Path(__file__).resolve().parent.parent / "personas"
STATE_FILE = CONFIG_DIR / ".active_persona"


def load_config() -> dict:
    if not CONFIG_FILE.exists():
        return {"plugins": {}}
    try:
        with open(CONFIG_FILE, "r", encoding="utf-8") as f:
            return json.load(f)
    except Exception as e:
        print(f"Error loading config.json: {e}", file=sys.stderr)
        return {"plugins": {}}


def save_config(config: dict) -> None:
    try:
        with open(CONFIG_FILE, "w", encoding="utf-8") as f:
            json.dump(config, f, indent=2)
            f.write("\n")
    except Exception as e:
        print(f"Error saving config.json: {e}", file=sys.stderr)


def get_available_personas() -> dict:
    personas = {}
    if not PERSONAS_DIR.exists():
        return personas
    for p_file in PERSONAS_DIR.glob("*.json"):
        try:
            with open(p_file, "r", encoding="utf-8") as f:
                data = json.load(f)
                personas[data.get("name", p_file.stem)] = data
        except Exception:
            continue
    return personas


def get_current_persona() -> str:
    if STATE_FILE.exists():
        try:
            val = STATE_FILE.read_text(encoding="utf-8").strip()
            if val:
                return val
        except Exception:
            pass
    return "core"


def set_current_persona(name: str) -> None:
    try:
        STATE_FILE.write_text(name.strip() + "\n", encoding="utf-8")
    except Exception as e:
        print(f"Warning: could not write active persona state: {e}", file=sys.stderr)


def ensure_catalog():
    CATALOG_DIR.mkdir(parents=True, exist_ok=True)


def switch_persona(persona_name: str, verbose: bool = True) -> bool:
    personas = get_available_personas()
    if persona_name not in personas:
        print(f"Error: Unknown persona '{persona_name}'. Available: {', '.join(personas.keys())}", file=sys.stderr)
        return False

    persona = personas[persona_name]
    config = load_config()
    if "plugins" not in config:
        config["plugins"] = {}

    # 1. Update plugins in config.json
    enabled_plugins = set(persona.get("enabled_plugins", []))
    disabled_plugins = set(persona.get("disabled_plugins", []))

    for plugin in enabled_plugins:
        if plugin not in config["plugins"]:
            config["plugins"][plugin] = {}
        config["plugins"][plugin]["enabled"] = True

    for plugin in disabled_plugins:
        if plugin in config["plugins"]:
            config["plugins"][plugin]["enabled"] = False
        else:
            config["plugins"][plugin] = {"enabled": False}

    save_config(config)

    # 2. Update active skills in ~/.gemini/config/skills/
    ensure_catalog()
    SKILLS_DIR.mkdir(parents=True, exist_ok=True)

    # Clean existing symlinks in SKILLS_DIR (preserve physical non-symlink directories like graphify)
    for item in SKILLS_DIR.iterdir():
        if item.is_symlink():
            try:
                item.unlink()
            except Exception as e:
                print(f"Warning: could not remove symlink {item}: {e}", file=sys.stderr)

    # Create symlinks for active_skills from CATALOG_DIR or ~/.agents/skills
    active_skills = persona.get("active_skills", [])
    linked_count = 0
    missing_skills = []

    for skill in active_skills:
        catalog_target = CATALOG_DIR / skill
        fallback_target = HOME / ".agents" / "skills" / skill

        target = None
        if catalog_target.exists():
            target = catalog_target
        elif fallback_target.exists():
            target = fallback_target

        if target:
            dest = SKILLS_DIR / skill
            try:
                if not dest.exists():
                    dest.symlink_to(target)
                linked_count += 1
            except Exception as e:
                print(f"Warning: failed to symlink {skill}: {e}", file=sys.stderr)
        else:
            missing_skills.append(skill)

    set_current_persona(persona_name)

    if verbose:
        print(f" switched to persona: \033[1;36m{persona.get('displayName', persona_name)}\033[0m ({persona_name})")
        print(f"   Description: {persona.get('description', '')}")
        print(f"   Plugins enabled : {', '.join(enabled_plugins)}")
        print(f"   Skills active   : {linked_count} loaded")
        if missing_skills:
            print(f"   Notice: {len(missing_skills)} skills not found in catalog ({', '.join(missing_skills[:3])}...)")
    return True


def list_personas() -> None:
    personas = get_available_personas()
    current = get_current_persona()

    print("\n\033[1;35mAntigravity Personas\033[0m")
    print("=" * 60)
    for name, data in sorted(personas.items()):
        is_cur = (name == current)
        marker = "\033[1;32m* (active)\033[0m" if is_cur else "  "
        print(f"{marker} \033[1;37m{name:<14}\033[0m : {data.get('displayName', name)}")
        print(f"     {data.get('description', '')}")
        print(f"     Plugins : {', '.join(data.get('enabled_plugins', []))}")
        print(f"     Skills  : {len(data.get('active_skills', []))} configured")
        print()


def show_current() -> None:
    current = get_current_persona()
    personas = get_available_personas()
    data = personas.get(current, {})

    print(f"Active Persona : \033[1;36m{data.get('displayName', current)}\033[0m ({current})")
    print(f"Description    : {data.get('description', 'N/A')}")
    
    # Active skills in directory
    active_skills = []
    if SKILLS_DIR.exists():
        for item in sorted(SKILLS_DIR.iterdir()):
            active_skills.append(item.name)
    print(f"Active Skills  ({len(active_skills)}): {', '.join(active_skills[:8])}{'...' if len(active_skills) > 8 else ''}")

    # Active plugins in config.json
    config = load_config()
    enabled_plugins = [k for k, v in config.get("plugins", {}).items() if v.get("enabled")]
    print(f"Active Plugins ({len(enabled_plugins)}): {', '.join(enabled_plugins)}")


def load_item(item_name: str) -> None:
    """Temporarily load an individual skill or plugin into the active session."""
    # Check if it's a plugin in ~/.gemini/config/plugins
    plugin_path = CONFIG_DIR / "plugins" / item_name
    if plugin_path.exists():
        config = load_config()
        if "plugins" not in config:
            config["plugins"] = {}
        config["plugins"][item_name] = {"enabled": True}
        save_config(config)
        print(f" Plugin enabled: \033[1;32m{item_name}\033[0m")
        return

    # Check if it's a skill
    catalog_target = CATALOG_DIR / item_name
    fallback_target = HOME / ".agents" / "skills" / item_name
    target = catalog_target if catalog_target.exists() else (fallback_target if fallback_target.exists() else None)

    if target:
        dest = SKILLS_DIR / item_name
        if not dest.exists():
            dest.symlink_to(target)
        print(f" Skill loaded: \033[1;32m{item_name}\033[0m")
        return

    print(f"Error: Could not find plugin or skill named '{item_name}'", file=sys.stderr)


def unload_item(item_name: str) -> None:
    """Unload an individual skill or plugin."""
    config = load_config()
    if item_name in config.get("plugins", {}):
        config["plugins"][item_name]["enabled"] = False
        save_config(config)
        print(f" Plugin disabled: \033[1;33m{item_name}\033[0m")

    dest = SKILLS_DIR / item_name
    if dest.is_symlink():
        dest.unlink()
        print(f" Skill unloaded: \033[1;33m{item_name}\033[0m")
    elif dest.exists() and not dest.is_symlink():
        print(f"Notice: '{item_name}' is a physical directory, not removing to protect data.", file=sys.stderr)


def main():
    parser = argparse.ArgumentParser(description="Antigravity Persona and Skill/Plugin Manager")
    subparsers = parser.add_subparsers(dest="command")

    subparsers.add_parser("list", help="List all available personas")
    subparsers.add_parser("current", help="Show current persona and active components")
    subparsers.add_parser("reset", help="Reset to minimal core persona")

    switch_parser = subparsers.add_parser("switch", help="Switch to a persona")
    switch_parser.add_argument("persona", help="Persona name (core, architect, fullstack, stitch, gcp-sre, adk-dev)")

    load_parser = subparsers.add_parser("load", help="Load an individual skill or plugin")
    load_parser.add_argument("name", help="Name of skill or plugin")

    unload_parser = subparsers.add_parser("unload", help="Unload an individual skill or plugin")
    unload_parser.add_argument("name", help="Name of skill or plugin")

    args = parser.parse_args()

    if args.command == "list":
        list_personas()
    elif args.command == "current":
        show_current()
    elif args.command == "switch":
        switch_persona(args.persona)
    elif args.command == "reset":
        switch_persona("core")
    elif args.command == "load":
        load_item(args.name)
    elif args.command == "unload":
        unload_item(args.name)
    else:
        show_current()


if __name__ == "__main__":
    main()
