// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package fsutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAtomicWriteFile(t *testing.T) {
	tmpDir := t.TempDir()
	targetFile := filepath.Join(tmpDir, "sub", "test.json")
	content := []byte(`{"hello": "world"}`)

	if err := AtomicWriteFile(targetFile, content, 0644); err != nil {
		t.Fatalf("AtomicWriteFile failed: %v", err)
	}

	read, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("reading written file: %v", err)
	}
	if string(read) != string(content) {
		t.Fatalf("got %s, want %s", string(read), string(content))
	}
}

func TestSymlinkManagement(t *testing.T) {
	tmpDir := t.TempDir()
	target1 := filepath.Join(tmpDir, "target1")
	target2 := filepath.Join(tmpDir, "target2")
	linkPath := filepath.Join(tmpDir, "link")

	if err := os.WriteFile(target1, []byte("one"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target2, []byte("two"), 0644); err != nil {
		t.Fatal(err)
	}

	// 1. Create symlink to target1
	if err := CreateOrUpdateSymlink(target1, linkPath); err != nil {
		t.Fatalf("CreateOrUpdateSymlink failed: %v", err)
	}

	isLink, err := IsSymlink(linkPath)
	if err != nil || !isLink {
		t.Fatalf("expected symlink, got isLink=%v, err=%v", isLink, err)
	}

	// 2. Update symlink to target2
	if err := CreateOrUpdateSymlink(target2, linkPath); err != nil {
		t.Fatalf("updating symlink failed: %v", err)
	}
	cur, _ := os.Readlink(linkPath)
	if cur != target2 {
		t.Fatalf("expected target2, got %s", cur)
	}

	// 3. Remove symlink
	removed, err := RemoveSymlink(linkPath)
	if err != nil || !removed {
		t.Fatalf("expected removed=true, got %v, err=%v", removed, err)
	}

	// 4. Protect regular directory
	physDir := filepath.Join(tmpDir, "physical_dir")
	if err := os.Mkdir(physDir, 0755); err != nil {
		t.Fatal(err)
	}
	removedPhys, err := RemoveSymlink(physDir)
	if err != nil || removedPhys {
		t.Fatalf("expected removedPhys=false, got %v, err=%v", removedPhys, err)
	}
	if _, err := os.Stat(physDir); os.IsNotExist(err) {
		t.Fatal("physical directory was deleted!")
	}
}
