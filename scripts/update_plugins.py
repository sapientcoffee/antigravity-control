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

"""Antigravity Plugin Version and Update Engine.

Audits plugins in ~/.gemini/config/plugins against their upstream GitHub
repositories and git remotes, reporting version diffs and providing one-step updates.
"""

import argparse
import json
import os
from pathlib import Path
import re
import subprocess
import sys
from typing import Dict, List, Optional, Tuple

HOME = Path.home()
PLUGINS_DIR = HOME / ".gemini" / "config" / "plugins"
CACHE_DIR = HOME / ".cache" / "antigravity-hub"
CACHE_FILE = CACHE_DIR / "update_audit.json"


def parse_repo_owner_name(url: str) -> Optional[Tuple[str, str]]:
    if not url:
        return None
    url = url.strip().rstrip("/")
    # Handle ssh or https git urls: https://github.com/owner/repo or git@github.com:owner/repo
    m = re.search(r"github\.com[/:]([A-Za-z0-9_.-]+)/([A-Za-z0-9_.-]+?)(?:\.git)?$", url)
    if m:
        return (m.group(1), m.group(2))
    return None


def get_git_info(path: Path) -> Optional[Dict[str, str]]:
    target = path.resolve() if path.is_symlink() else path
    git_dir = target / ".git"
    if not git_dir.exists():
        return None
    try:
        commit = subprocess.check_output(
            ["git", "-C", str(target), "rev-parse", "HEAD"],
            text=True, stderr=subprocess.DEVNULL
        ).strip()
        remote_url = ""
        try:
            remote_url = subprocess.check_output(
                ["git", "-C", str(target), "config", "--get", "remote.origin.url"],
                text=True, stderr=subprocess.DEVNULL
            ).strip()
        except Exception:
            pass
        return {"commit": commit, "remote_url": remote_url, "path": str(target)}
    except Exception:
        return None


def get_remote_git_head(remote_url: str) -> Optional[str]:
    try:
        output = subprocess.check_output(
            ["git", "ls-remote", remote_url, "HEAD"],
            text=True, stderr=subprocess.DEVNULL, timeout=4
        ).strip()
        if output:
            return output.split()[0]
    except Exception:
        pass
    return None


def get_remote_github_latest(owner: str, repo: str) -> Optional[Dict[str, str]]:
    """Query GitHub via gh CLI for latest release tag or commit."""
    # 1. Try gh api releases/latest
    try:
        res = subprocess.run(
            ["gh", "api", f"repos/{owner}/{repo}/releases/latest"],
            capture_output=True, text=True, timeout=3
        )
        if res.returncode == 0:
            data = json.loads(res.stdout)
            tag = data.get("tag_name", "").lstrip("v")
            if tag:
                return {"type": "release", "latest": tag}
    except Exception:
        pass

    # 2. Try gh api tags
    try:
        res = subprocess.run(
            ["gh", "api", f"repos/{owner}/{repo}/tags", "--jq", ".[0].name"],
            capture_output=True, text=True, timeout=3
        )
        if res.returncode == 0 and res.stdout.strip():
            tag = res.stdout.strip().lstrip("v")
            return {"type": "tag", "latest": tag}
    except Exception:
        pass

    # 3. Fallback: query HEAD commit
    try:
        res = subprocess.run(
            ["gh", "api", f"repos/{owner}/{repo}/commits/HEAD", "--jq", ".sha"],
            capture_output=True, text=True, timeout=3
        )
        if res.returncode == 0 and res.stdout.strip():
            return {"type": "commit", "latest": res.stdout.strip()[:7]}
    except Exception:
        pass

    return None


def audit_plugins() -> List[Dict]:
    results = []
    if not PLUGINS_DIR.exists():
        return results

    for plugin_path in sorted(PLUGINS_DIR.iterdir()):
        if not plugin_path.is_dir() and not (plugin_path.is_symlink() and plugin_path.resolve().is_dir()):
            continue

        manifest_file = plugin_path / "plugin.json"
        name = plugin_path.name
        local_version = "unknown"
        repo_url = ""
        description = ""

        if manifest_file.exists():
            try:
                with open(manifest_file, "r", encoding="utf-8") as f:
                    data = json.load(f)
                    name = data.get("name", name)
                    local_version = str(data.get("version", "unknown"))
                    repo_url = data.get("repository", "")
                    description = data.get("description", "")
            except Exception:
                pass

        git_info = get_git_info(plugin_path)
        is_symlink = plugin_path.is_symlink()
        is_local_git = git_info is not None

        remote_version = "n/a"
        status = "UP-TO-DATE"
        check_details = ""

        repo_parsed = parse_repo_owner_name(repo_url)

        # Check via Git
        if is_local_git and git_info.get("remote_url"):
            remote_head = get_remote_git_head(git_info["remote_url"])
            local_commit = git_info["commit"][:7]
            if remote_head:
                remote_short = remote_head[:7]
                remote_version = remote_short
                if local_commit != remote_short:
                    status = "OUTDATED"
                    check_details = f"commit {local_commit} -> {remote_short}"
                else:
                    status = "UP-TO-DATE"
                    check_details = f"git @ {local_commit}"
            else:
                remote_version = local_commit
                check_details = f"local git @ {local_commit}"

        elif repo_parsed:
            owner, repo = repo_parsed
            remote_info = get_remote_github_latest(owner, repo)
            if remote_info:
                remote_version = remote_info["latest"]
                clean_local = local_version.lstrip("v")
                clean_remote = remote_version.lstrip("v")
                if clean_local != "unknown" and clean_local != clean_remote:
                    status = "OUTDATED"
                    check_details = f"{clean_local} -> {clean_remote}"
                else:
                    status = "UP-TO-DATE"
                    check_details = f"{remote_info['type']}: {clean_remote}"

        results.append({
            "name": name,
            "path": str(plugin_path),
            "is_symlink": is_symlink,
            "is_git": is_local_git,
            "local_version": local_version,
            "remote_version": remote_version,
            "status": status,
            "repo_url": repo_url,
            "details": check_details
        })

    # Cache results
    try:
        CACHE_DIR.mkdir(parents=True, exist_ok=True)
        with open(CACHE_FILE, "w", encoding="utf-8") as f:
            json.dump(results, f, indent=2)
    except Exception:
        pass

    return results


def print_table(results: List[Dict]):
    print("\n\033[1;35mAntigravity Plugin Version Audit\033[0m")
    print("=" * 86)
    print(f"{'PLUGIN':<28} {'LOCAL VER':<14} {'REMOTE':<14} {'STATUS':<12} {'DETAILS':<14}")
    print("-" * 86)

    for item in results:
        status_color = "\033[1;32m" if item["status"] == "UP-TO-DATE" else "\033[1;31m"
        print(
            f"{item['name']:<28} "
            f"{item['local_version']:<14} "
            f"{item['remote_version']:<14} "
            f"{status_color}{item['status']:<12}\033[0m "
            f"{item['details']:<14}"
        )
    print("-" * 86)

    outdated = [r for r in results if r["status"] == "OUTDATED"]
    if outdated:
        print(f"\033[1;33m💡 {len(outdated)} plugin(s) have updates available!\033[0m Run \033[1;36magyctl update [name|all]\033[0m to update.\n")
    else:
        print("\033[1;32m All Antigravity plugins are up-to-date.\033[0m\n")


def print_markdown(results: List[Dict]):
    outdated = [r for r in results if r["status"] == "OUTDATED"]
    if not outdated:
        return

    print("### 🔄 Antigravity Plugin Updates Available")
    print(f"> Found **{len(outdated)}** plugin(s) with newer versions upstream:")
    print("")
    for item in outdated:
        print(f"- **`{item['name']}`**: `{item['local_version']}` ➔ `{item['remote_version']}` ({item.get('details', '')})")
    print("")
    print("Run `agyctl update` to update.")


def update_plugin(item: Dict) -> bool:
    name = item["name"]
    path = Path(item["path"])
    target = path.resolve() if path.is_symlink() else path

    print(f"Updating plugin: \033[1;36m{name}\033[0m...")

    # If it's a git repo or symlink to git repo, perform git pull
    if (target / ".git").exists():
        try:
            res = subprocess.run(
                ["git", "-C", str(target), "pull", "--ff-only"],
                capture_output=True, text=True
            )
            if res.returncode == 0:
                print(f"  Successfully pulled latest commits for {name}.")
                return True
            else:
                print(f"  git pull failed: {res.stderr.strip()}", file=sys.stderr)
        except Exception as e:
            print(f"  Error updating {name}: {e}", file=sys.stderr)
            return False

    # Non-git plugin with a known GitHub repository: clone or fetch
    repo_url = item.get("repo_url")
    if repo_url:
        print(f"  Plugin is not a git checkout. Re-cloning latest from {repo_url}...")
        try:
            backup_dir = target.with_name(target.name + ".bak")
            if backup_dir.exists():
                shutil.rmtree(backup_dir)
            target.rename(backup_dir)

            res = subprocess.run(["git", "clone", "--depth", "1", repo_url, str(target)], capture_output=True, text=True)
            if res.returncode == 0:
                shutil.rmtree(backup_dir)
                print(f"  Successfully updated {name} from remote repository.")
                return True
            else:
                backup_dir.rename(target)
                print(f"  Clone failed: {res.stderr.strip()}", file=sys.stderr)
        except Exception as e:
            print(f"  Error: {e}", file=sys.stderr)

    return False


def main():
    parser = argparse.ArgumentParser(description="Antigravity Plugin Version and Update Engine")
    subparsers = parser.add_subparsers(dest="command")

    subparsers.add_parser("list", help="List installed plugins and versions")
    check_parser = subparsers.add_parser("check", help="Check plugins for upstream updates")
    check_parser.add_argument("--json", action="store_true", help="Output as JSON")
    check_parser.add_argument("--markdown", action="store_true", help="Output as Markdown (for SessionStart hook)")
    check_parser.add_argument("--quiet", action="store_true", help="Quiet output, exit code 1 if outdated")

    update_parser = subparsers.add_parser("update", help="Update outdated plugins")
    update_parser.add_argument("name", nargs="?", default="all", help="Plugin name or 'all'")

    args = parser.parse_args()

    results = audit_plugins()

    if args.command == "check":
        if args.json:
            print(json.dumps(results, indent=2))
        elif args.markdown:
            print_markdown(results)
        elif not args.quiet:
            print_table(results)
        outdated_count = len([r for r in results if r["status"] == "OUTDATED"])
        sys.exit(1 if outdated_count > 0 else 0)

    elif args.command == "update":
        target = args.name
        to_update = [r for r in results if (target == "all" or r["name"] == target) and r["status"] == "OUTDATED"]
        if not to_update:
            print(f"No outdated plugins found matching '{target}'.")
            return
        success = 0
        for item in to_update:
            if update_plugin(item):
                success += 1
        print(f"\nUpdated {success}/{len(to_update)} plugin(s).")

    else:
        print_table(results)


if __name__ == "__main__":
    main()
