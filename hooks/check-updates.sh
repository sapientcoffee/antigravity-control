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

CACHE_DIR="$HOME/.cache/antigravity-hub"
CACHE_OUT="$CACHE_DIR/cached_notice.md"
TS_FILE="$CACHE_DIR/last_check_ts"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENGINE="$SCRIPT_DIR/../scripts/update_plugins.py"

mkdir -p "$CACHE_DIR"

NOW=$(date +%s)
LAST_CHECK=0
if [ -f "$TS_FILE" ]; then
    LAST_CHECK=$(cat "$TS_FILE" 2>/dev/null || echo 0)
fi

AGE=$((NOW - LAST_CHECK))
# Check every 6 hours (21600s)
TTL=21600

if [ $AGE -gt $TTL ] || [ ! -f "$CACHE_OUT" ]; then
    python3 "$ENGINE" check --markdown > "$CACHE_OUT" 2>/dev/null
    echo "$NOW" > "$TS_FILE"
fi

# If there is cached markdown notice, output it to stdout for AGY session injection
if [ -f "$CACHE_OUT" ] && [ -s "$CACHE_OUT" ]; then
    cat "$CACHE_OUT"
fi
