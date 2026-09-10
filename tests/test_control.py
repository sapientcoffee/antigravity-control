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

import json
from pathlib import Path
import subprocess
import sys
import unittest

REPO_DIR = Path(__file__).resolve().parent.parent
BIN_DIR = REPO_DIR / "bin"
SCRIPTS_DIR = REPO_DIR / "scripts"
PERSONAS_DIR = REPO_DIR / "personas"


class TestAntigravityControl(unittest.TestCase):

    def test_persona_files_valid(self):
        """All persona files should be valid JSON and contain required keys."""
        persona_files = list(PERSONAS_DIR.glob("*.json"))
        self.assertGreaterEqual(len(persona_files), 5)

        for p_file in persona_files:
            with open(p_file, "r", encoding="utf-8") as f:
                data = json.load(f)
                self.assertIn("name", data)
                self.assertIn("displayName", data)
                self.assertIn("enabled_plugins", data)
                self.assertIn("disabled_plugins", data)
                self.assertIn("active_skills", data)
                self.assertIn("antigravity-control", data["enabled_plugins"])

    def test_agyctl_list(self):
        """agyctl list should execute with exit code 0."""
        cmd = [str(BIN_DIR / "agyctl"), "list"]
        res = subprocess.run(cmd, capture_output=True, text=True)
        self.assertEqual(res.returncode, 0)
        self.assertIn("core", res.stdout)
        self.assertIn("architect", res.stdout)
        self.assertIn("fullstack", res.stdout)
        self.assertIn("stitch", res.stdout)

    def test_agyctl_check(self):
        """agyctl check should execute cleanly."""
        cmd = [str(BIN_DIR / "agyctl"), "check", "--json"]
        res = subprocess.run(cmd, capture_output=True, text=True)
        # returncode is 0 (all up to date) or 1 (updates available)
        self.assertIn(res.returncode, [0, 1])
        data = json.loads(res.stdout)
        self.assertIsInstance(data, list)
        self.assertGreater(len(data), 0)
        self.assertIn("name", data[0])
        self.assertIn("status", data[0])

    def test_check_updates_hook(self):
        """check-updates.sh hook script should run quickly."""
        hook_script = REPO_DIR / "hooks" / "check-updates.sh"
        res = subprocess.run([str(hook_script)], capture_output=True, text=True)
        self.assertEqual(res.returncode, 0)


if __name__ == "__main__":
    unittest.main()
