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
	"fmt"
	"os"
	"path/filepath"
)

// EnsureDir creates the directory and all parent directories if they don't exist.
func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// IsSymlink reports whether path is a symbolic link.
func IsSymlink(path string) (bool, error) {
	fi, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return fi.Mode()&os.ModeSymlink != 0, nil
}

// AtomicWriteFile writes data to a temporary file in the same directory, then renames it
// atomically to the destination path.
func AtomicWriteFile(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	if err := EnsureDir(dir); err != nil {
		return fmt.Errorf("creating directory %s: %w", dir, err)
	}

	tmpFile, err := os.CreateTemp(dir, ".tmp-"+filepath.Base(path)+"-*")
	if err != nil {
		return fmt.Errorf("creating temp file in %s: %w", dir, err)
	}
	tmpName := tmpFile.Name()
	defer func() {
		_ = os.Remove(tmpName)
	}()

	if _, err := tmpFile.Write(data); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("writing to temp file: %w", err)
	}

	if err := tmpFile.Chmod(perm); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("setting permissions on temp file: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("closing temp file: %w", err)
	}

	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("atomic rename from %s to %s: %w", tmpName, path, err)
	}

	return nil
}

// RemoveSymlink removes the file at path only if it is a symbolic link.
// If the path does not exist, it returns false, nil.
// If the path exists and is NOT a symlink, it leaves the file untouched and returns false, nil.
// If it was a symlink and was removed, it returns true, nil.
func RemoveSymlink(path string) (bool, error) {
	isLink, err := IsSymlink(path)
	if err != nil {
		return false, err
	}
	if !isLink {
		return false, nil
	}
	if err := os.Remove(path); err != nil {
		return false, err
	}
	return true, nil
}

// CreateOrUpdateSymlink creates a symlink pointing to target at linkPath.
// If a symlink already exists at linkPath, it is cleanly replaced.
// If a non-symlink file/directory exists at linkPath, an error is returned to prevent data loss.
func CreateOrUpdateSymlink(target, linkPath string) error {
	fi, err := os.Lstat(linkPath)
	if err == nil {
		// Target already exists
		if fi.Mode()&os.ModeSymlink != 0 {
			// It is a symlink, check where it points
			current, readErr := os.Readlink(linkPath)
			if readErr == nil && current == target {
				return nil // already points to correct target
			}
			// Remove existing symlink to replace
			if err := os.Remove(linkPath); err != nil {
				return fmt.Errorf("removing existing symlink %s: %w", linkPath, err)
			}
		} else {
			return fmt.Errorf("path %s already exists and is not a symlink; preserving to protect data", linkPath)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("checking path %s: %w", linkPath, err)
	}

	if err := EnsureDir(filepath.Dir(linkPath)); err != nil {
		return fmt.Errorf("ensuring parent directory for %s: %w", linkPath, err)
	}

	if err := os.Symlink(target, linkPath); err != nil {
		return fmt.Errorf("creating symlink from %s to %s: %w", linkPath, target, err)
	}

	return nil
}
